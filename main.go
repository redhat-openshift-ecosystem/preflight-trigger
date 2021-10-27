package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/openshift/ci-tools/pkg/api"
	"github.com/openshift/ci-tools/pkg/steps/utils"
	"github.com/openshift/ci-tools/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	pjapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
	pjclientset "k8s.io/test-infra/prow/client/clientset/versioned"
	prowconfig "k8s.io/test-infra/prow/config"
	configflagutil "k8s.io/test-infra/prow/flagutil/config"
	"k8s.io/test-infra/prow/gcsupload"
	"k8s.io/test-infra/prow/interrupts"
	"k8s.io/test-infra/prow/pjutil"
	"k8s.io/test-infra/prow/pod-utils/decorate"
	"k8s.io/test-infra/prow/pod-utils/downwardapi"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Constants for flags
const (
	hiddenOption = "hidden"
	jobConfigPathOption = "job-config-path"
	jobNameOption = "job-name"
	ocpVersionOption = "ocp-version"
	outputFilePathOption = "output-path"
	prowConfigPathOption = "prow-config-path"
	releaseImageRefOption = "release-image-ref"
	testAssetOption = "test-asset"
	assetTypeOption = "asset-type"
	pfltLogLevelOption = "pflt-log-level"
	pfltLogFileOption = "pflt-log-file"
	pfltArtifactsOption = "pflt-artifacts"
	pfltNamespaceOption = "pflt-namespace"
	pfltServiceAccountOption = "pflt-service-account"
	pfltIndexImageOption = "pflt-index-image"
	pfltDockerConfigOption = "pflt-docker-config"
)

var fileSystem = afero.NewOsFs()
var fs = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
var o options

type options struct {
	hidden bool
	jobName string
	prowconfig configflagutil.ConfigOptions
	ocpVersion string
	outputPath string
	releaseImageRef string
	testAsset string
	assetType string
	pfltLogLevel string
	pfltLogFile string
	pfltArtifacts string
	pfltNamespace string
	pfltServiceAccount string
	pfltIndexImage string
	pfltDockerConfig string
	dryRun bool
}

type prowjobResult struct {
	Status pjapi.ProwJobState `json:"status"`
	ArtifactsURL string `json:"prowjob_artifacts_url"`
	URL string `json:"prowjob_url"`
}

type jobResult interface {
	toJSON() ([]byte, error)
}

func (p *prowjobResult) toJSON() ([]byte, error) {
	return json.MarshalIndent(p, "", "    ")
}

func (o *options) gatherOptions() {
	o.prowconfig.ConfigPathFlagName = prowConfigPathOption
	o.prowconfig.JobConfigPathFlagName = jobConfigPathOption
	fs.StringVar(&o.jobName, jobNameOption, "", "Name of the periodic job to manually trigger")
	fs.StringVar(&o.ocpVersion, ocpVersionOption, "", "Version of OCP to use; 4.x or higher")
	fs.StringVar(&o.outputPath, outputFilePathOption, "", "File to store JSON returned from job submission")
	fs.StringVar(&o.releaseImageRef, releaseImageRefOption, "", "Release payload image to use for OCP deployment.")
	fs.StringVar(&o.testAsset, testAssetOption, "", "Provided to preflight check documentation")
	fs.StringVar(&o.assetType, assetTypeOption, "", "Provided to preflight check documentation")
	fs.StringVar(&o.pfltLogLevel, pfltLogLevelOption, "", "Provided to preflight check documentation")
	fs.StringVar(&o.pfltLogFile, pfltLogFileOption, "", "Provided to preflight check documentation")
	fs.StringVar(&o.pfltArtifacts, pfltArtifactsOption, "", "Provided to preflight check documentation")
	fs.StringVar(&o.pfltNamespace, pfltNamespaceOption, "", "Provided to preflight check documentation")
	fs.StringVar(&o.pfltServiceAccount, pfltServiceAccountOption, "", "Provided to preflight check documentation")
	fs.StringVar(&o.pfltIndexImage, pfltIndexImageOption, "", "Provided to preflight check documentation")
	fs.StringVar(&o.pfltDockerConfig, pfltDockerConfigOption, "", "Provided to preflight check documentation")
	fs.BoolVar(&o.dryRun, "dry-run", false, "Display the job YAML without submitting the job to Prow")
	fs.BoolVar(&o.hidden, "hidden", false, "Hide job from Prow Deck output")
	o.prowconfig.AddFlags(fs)
}

func (o options) validateOptions() error {
	afs := afero.Afero{ Fs: fileSystem }

	exists, _ := afs.Exists(o.prowconfig.JobConfigPath)
	if !exists {
		return fmt.Errorf("validation error: job config path does not exist (%s)", o.prowconfig.JobConfigPath)
	}

	if o.jobName == "" {
		return fmt.Errorf("%s flag is required", jobNameOption)
	}

	if o.ocpVersion == "" {
		return fmt.Errorf("%s flag is required", ocpVersionOption)
	}

	if o.testAsset == "" {
		return fmt.Errorf("%s flag is required", testAssetOption)
	}

	if o.assetType == "" {
		return fmt.Errorf("%s flag is required", assetTypeOption)
	}

	if !strings.HasPrefix(o.ocpVersion, "4") {
		return fmt.Errorf("%s must be 4.x or higher", ocpVersionOption)
	}

	exists, _ = afs.Exists(o.prowconfig.ConfigPath)
	if !exists {
		return fmt.Errorf("validation error: prow config path does not exist (%s)", o.prowconfig.ConfigPath)
	}

	if !o.dryRun {
		if o.outputPath == "" {
			return fmt.Errorf("%s flag is required", outputFilePathOption)
		}
		exists, _ = afs.Exists(filepath.Dir(o.outputPath))
		if !exists {
			return fmt.Errorf("validation error: output file path does not exist (%s)", o.outputPath)
		}
	}

	return nil
}

func getPeriodicJob(jobName string, config *prowconfig.Config) (*pjapi.ProwJob, error) {
	var selectedJob *prowconfig.Periodic
	for _, job := range config.AllPeriodics() {
		if job.Name == jobName {
			selectedJob = &job
			break
		}
	}

	if selectedJob == nil {
		return nil, fmt.Errorf("failed to find the job: %s", jobName)
	}

	prowjob := pjutil.NewProwJob(pjutil.PeriodicSpec(*selectedJob), nil, nil)
	return &prowjob, nil
}

// returns the artifacts URL for the given job
func getJobArtifactsURL(prowJob *pjapi.ProwJob, config *prowconfig.Config) string {
	var identifier string
	if prowJob.Spec.Refs != nil {
		identifier = fmt.Sprintf("%s/%s", prowJob.Spec.Refs.Org, prowJob.Spec.Refs.Repo)
	} else {
		identifier = fmt.Sprintf("%s/%s", prowJob.Spec.ExtraRefs[0].Org, prowJob.Spec.ExtraRefs[0].Repo)
	}
	spec := downwardapi.NewJobSpec(prowJob.Spec, prowJob.Status.BuildID, prowJob.Name)
	gcsConfig := config.Plank.GuessDefaultDecorationConfig(identifier, prowJob.Spec.Cluster).GCSConfiguration
	jobBasePath, _, _ := gcsupload.PathsForJob(gcsConfig, &spec, "")
	return fmt.Sprintf("%s%s/%s",
		config.Deck.Spyglass.GCSBrowserPrefix,
		gcsConfig.Bucket,
		jobBasePath,
	)
}

// Calls toJSON method on a jobResult type and writes it to the output path
func writeResultOutput(prowjobResult jobResult, outputPath string) error {
	j, err := prowjobResult.toJSON()
	if err != nil {
		logrus.Error("Unable to marshal prowjob result to JSON")
		return err
	}

	afs := afero.Afero{Fs: fileSystem}
	err = afs.WriteFile(outputPath, j, 0755)
	if err != nil {
		logrus.WithField("output path", outputPath).Error("error writing to output file")
		return err
	}

	return nil
}

// appendMultiStageParams passes all the OO_ params to ci-operator as multi-stage-params.
func appendMultiStageParams(podSpec *v1.PodSpec, params map[string]string) {
	// for execution purposes, the order isn't super important, but in order to allow for consistent test verification we need
	// to sort the params.
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		podSpec.Containers[0].Args = append(podSpec.Containers[0].Args, fmt.Sprintf("--multi-stage-param=%s=\"%s\"", key, params[key]))
	}
}

// appendMultiStageParams passes image dependency overrides to ci-operator
func appendMultiStageDepOverrides(podSpec *v1.PodSpec, overrides map[string]string) {
	keys := make([]string, 0, len(overrides))
	for key := range overrides {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		podSpec.Containers[0].Args = append(podSpec.Containers[0].Args, fmt.Sprintf("--dependency-override-param=%s=\"%s\"", key, overrides[key]))
	}
}

func main() {
	o.gatherOptions()
	err := fs.Parse(os.Args[1:])
	if err != nil {
		logrus.WithError(err).Fatal("error parsing flag set")
	}

	err = o.validateOptions()
	if err != nil {
		logrus.WithError(err).Fatal("invalid options")
	}

	go func() {
		interrupts.WaitForGracefulShutdown()
		os.Exit(128)
	}()

	configAgent, err := o.prowconfig.ConfigAgent()
	if err != nil {
		logrus.WithError(err).Fatal("failed to read prow configuration")
	}
	config := configAgent.Config()
	prowjob, err := getPeriodicJob(o.jobName, config)
	if err != nil {
		logrus.WithField("job-name", o.jobName).Fatal(err)
	}

	jobparams := make(map[string]string)

	params := map[string]string{
		"PFLT_LOGLEVEL": o.pfltLogLevel,
		"PFLT_LOGFILE": o.pfltLogFile,
		"PFLT_ARTIFACTS": o.pfltArtifacts,
		"PFLT_NAMESPACE": o.pfltNamespace,
		"PFLT_SERVICEACCOUNT": o.pfltServiceAccount,
		"PFLT_INDEXIMAGE": o.pfltIndexImage,
		"PFLT_DOCKERCONFIG": o.pfltDockerConfig,
		"TEST_ASSET": o.testAsset,
		"ASSET_TYPE": o.assetType,
	}

	for k, v := range params {
		if v == "" {
			continue
		} else {
			jobparams[k] = v
		}
	}

	appendMultiStageParams(prowjob.Spec.PodSpec, jobparams)

	envvars := map[string]string{
		"CLUSTER_TYPE": "aws", // this would be configurable?
		"OCP_VERSION": o.ocpVersion,
	}
	if o.releaseImageRef != "" {
		envvars[utils.ReleaseImageEnv(api.LatestReleaseName)] = o.releaseImageRef
	}

	var keys []string
	for key := range envvars {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	// TODO: disabling args for now so echo-test can run
	input := strings.Builder{}
	input.WriteString("--input-hash=")
	for _, key := range keys {
		input.WriteString(key)
		input.WriteString(envvars[key])
	}
	prowjob.Spec.PodSpec.Containers[0].Args = append(prowjob.Spec.PodSpec.Containers[0].Args, input.String())
	prowjob.Spec.PodSpec.Containers[0].Env = append(prowjob.Spec.PodSpec.Containers[0].Env, decorate.KubeEnv(envvars)...)

	if o.hidden {
		prowjob.Spec.Hidden = true
	}

	if o.dryRun {
		jobAsYaml, err := yaml.Marshal(prowjob)
		if err != nil {
			logrus.WithError(err).Fatal("failed to marshal prowjob to yaml")
		}
		fmt.Println(string(jobAsYaml))
		os.Exit(0)
	}

	logrus.Info("getting cluster config")
	clusterConfig, err := util.LoadClusterConfig()
	if err != nil {
		logrus.WithError(err).Fatal("failed to load cluster configuration")
	}

	logrus.WithFields(pjutil.ProwJobFields(prowjob)).Info("submitting a new prowjob")
	pjcset, err := pjclientset.NewForConfig(clusterConfig)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create prowjob clientset")
	}

	pjclient := pjcset.ProwV1().ProwJobs(config.ProwJobNamespace)

	logrus.WithFields(pjutil.ProwJobFields(prowjob)).Info("submitting a new prowjob")
	created, err := pjclient.Create(context.TODO(), prowjob, metav1.CreateOptions{})
	if err != nil {
		logrus.WithError(err).Fatal("failed to submit the prowjob")
	}

	logger := logrus.WithFields(pjutil.ProwJobFields(created))
	logger.Info("submitted the prowjob, waiting for its result")

	selector := fields.SelectorFromSet(map[string]string{"metadata.name": created.Name})

	for {
		var w watch.Interface
		if err = wait.ExponentialBackoff(wait.Backoff{Steps: 10, Duration: 10 * time.Second, Factor: 2}, func() (bool, error) {
			var err2 error
			w, err2 = pjclient.Watch(interrupts.Context(), metav1.ListOptions{FieldSelector: selector.String()})
			if err2 != nil {
				logrus.Error(err2)
				return false, nil
			}
			return true, nil
		}); err != nil {
			logrus.WithError(err).Fatal("failed to create watch for ProwJobs")
		}

		for event := range w.ResultChan() {
			prowJob, ok := event.Object.(*pjapi.ProwJob)
			if !ok {
				logrus.WithField("object-type", fmt.Sprintf("%T", event.Object)).Fatal("received an unexpected object from Watch")
			}

			prowJobArtifactsURL := getJobArtifactsURL(prowJob, config)

			switch prowJob.Status.State {
			case pjapi.FailureState, pjapi.AbortedState, pjapi.ErrorState:
				pjr := &prowjobResult{
					Status:       prowJob.Status.State,
					ArtifactsURL: prowJobArtifactsURL,
					URL:          prowJob.Status.URL,
				}
				fmt.Printf("%+v\n", pjr)
				err = writeResultOutput(pjr, o.outputPath)
				if err != nil {
					logrus.Error("Unable to write prowjob result to file")
				}
				logrus.Fatal("job failed")
			case pjapi.SuccessState:
				pjr := &prowjobResult{
					Status:       prowJob.Status.State,
					ArtifactsURL: prowJobArtifactsURL,
					URL:          prowJob.Status.URL,
				}
				err = writeResultOutput(pjr, o.outputPath)
				fmt.Printf("%+v\n", pjr)
				if err != nil {
					logrus.Error("Unable to write prowjob result to file")
				}
				logrus.Info("job succeeded")
				os.Exit(0)
			}
		}
	}
}
