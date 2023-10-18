package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	pjapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
	pjclient "k8s.io/test-infra/prow/client/clientset/versioned"
	prowconfig "k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/gcsupload"
	"k8s.io/test-infra/prow/pjutil"
	"k8s.io/test-infra/prow/pod-utils/decorate"
	"k8s.io/test-infra/prow/pod-utils/downwardapi"
)

type prowjobResult struct {
	ArtifactsURL string             `json:"prowjob_artifacts_url"`
	Status       pjapi.ProwJobState `json:"status"`
	URL          string             `json:"prowjob_url"`
}

type jobResult interface {
	toJSON() ([]byte, error)
}

func (p *prowjobResult) toJSON() ([]byte, error) {
	return json.MarshalIndent(p, "", "    ")
}

func CreateProwJobManifest(name string, config *prowconfig.Config) (*pjapi.ProwJob, error) {
	var periodicjob *prowconfig.Periodic
	for _, periodic := range config.AllPeriodics() {
		if periodic.Name == name {
			periodicjob = &periodic
			break
		}
	}

	if periodicjob == nil {
		return nil, &PreflightTriggerCustomError{Message: "Unable to get periodic job", Err: errors.New("job with name " + name + " not found")}
	}

	jobmanifest := pjutil.NewProwJob(pjutil.PeriodicSpec(*periodicjob), nil, nil)
	return &jobmanifest, nil
}

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
func writeResultOutput(pjr jobResult, outputPath string) error {
	pjrjson, err := pjr.toJSON()
	if err != nil {
		log.Fatal("Unable to marshal prowjob result to JSON")
		return err
	}

	err = os.WriteFile(outputPath, pjrjson, 0o755)
	if err != nil {
		log.Fatalf("Error writing result to file: %v", err)
		return err
	}

	return nil
}

func AppendMultiStageParams(podspec *corev1.PodSpec, params map[string]string) {
	// For execution purposes, the order isn't super important, but in order to allow for
	// consistent test verification we need to sort the params.
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if params[key] != "" {
			podspec.Containers[0].Args = append(podspec.Containers[0].Args, fmt.Sprintf("--multi-stage-param=%s=%s", key, params[key]))
		}
	}
}

func SetInputHash(podspec *corev1.PodSpec, clustertype, ocpversion string) {
	envvars := map[string]string{"CLUSTER_TYPE": clustertype, "OCP_VERSION": ocpversion}
	inputhash := strings.Builder{}
	inputhash.WriteString("--input-hash=")
	inputhash.WriteString("CLUSTER_TYPE" + clustertype)
	inputhash.WriteString("OCP_VERSION" + ocpversion)
	podspec.Containers[0].Args = append(podspec.Containers[0].Args, inputhash.String())
	podspec.Containers[0].Env = append(podspec.Containers[0].Env, decorate.KubeEnv(envvars)...)
}

func ProwJobSuccess(pj *pjapi.ProwJob, config *prowconfig.Config, output string) {
	prowJobArtifactsURL := getJobArtifactsURL(pj, config)
	pjr := &prowjobResult{
		Status:       pj.Status.State,
		ArtifactsURL: prowJobArtifactsURL,
		URL:          pj.Status.URL,
	}
	err := writeResultOutput(pjr, output)
	fmt.Printf("%+v\n", pjr)
	if err != nil {
		log.Fatal("Unable to write prowjob result to file")
	}
	log.Println("job succeeded")
	os.Exit(0)
}

func ProwJobFailure(pj *pjapi.ProwJob, config *prowconfig.Config, output string) {
	prowJobArtifactsURL := getJobArtifactsURL(pj, config)
	pjr := &prowjobResult{
		Status:       pj.Status.State,
		ArtifactsURL: prowJobArtifactsURL,
		URL:          pj.Status.URL,
	}
	fmt.Printf("%+v\n", pjr)
	err := writeResultOutput(pjr, output)
	if err != nil {
		log.Fatal("Unable to write prowjob result to file")
	}
	log.Fatal("job failed")
}

func ProwJobWatcher(namespace string, pjcs *pjclient.Clientset, selector string) (watch.Interface, error) {
	ctx := context.Background()
	var watcher watch.Interface

	err := wait.ExponentialBackoff(wait.Backoff{Steps: 5, Duration: 5 * time.Second, Factor: 5, Cap: 30 * time.Second},
		func() (bool, error) {
			var err error

			watcher, err = pjcs.ProwV1().ProwJobs(namespace).Watch(ctx, metav1.ListOptions{FieldSelector: selector})
			if err != nil {
				log.Fatalf("%v", err)
			}

			return true, nil
		})
	if err != nil {
		log.Fatalf("%v", err)
	}

	return watcher, nil
}

// appendMultiStageParams passes image dependency overrides to ci-operator
/*func appendMultiStageDepOverrides(podSpec *v1.PodSpec, overrides map[string]string) {
	keys := make([]string, 0, len(overrides))
	for key := range overrides {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		podSpec.Containers[0].Args = append(podSpec.Containers[0].Args, fmt.Sprintf("--dependency-override-param=%s=\"%s\"", key, overrides[key]))
	}
}*/
