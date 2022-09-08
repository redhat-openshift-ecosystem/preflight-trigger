/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"github.com/ghodss/yaml"
	"github.com/openshift/ci-tools/pkg/util"
	. "github.com/redhat-openshift-ecosystem/preflight-trigger/internal"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	pjapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
	pjclient "k8s.io/test-infra/prow/client/clientset/versioned"
	"log"
	"os"
	"strings"
	"time"
)

// jobCmd represents the job command
var jobCmd = &cobra.Command{
	Use:    "job",
	Short:  "On-demand ProwJob within the openshift-ci infrastructure",
	Long:   ``,
	PreRun: jobPreRun,
	Run:    jobRun,
}

func init() {
	createCmd.AddCommand(jobCmd)
	jobCmd.Flags().StringVarP(&CommandFlags.CIEnvironment, "environment", "", "common", "Set the environment to use; can be one of [common, preprod, prod]")
}

func validateJobFlags() {
	log.Println("Validating job flags")

	// Check that required flags are not set to their zero values
	requiredFlags := map[string]string{
		"AssetType":      "--asset-type",
		"OcpVersion":     "--ocp-version",
		"PfltIndexImage": "--pflt-index-image",
		"PfltLogLevel":   "--pflt-log-level",
		"TestAsset":      "--test-asset",
	}

	for flag, tag := range requiredFlags {
		value := CommandFlags.Get(flag)
		if value.IsZero() {
			log.Printf("%s is required", tag)
			log.Fatalf("Required flags for job: %v", requiredFlags)
		}
	}

	if !strings.HasPrefix(CommandFlags.OcpVersion, "4") {
		log.Fatalln("Only OCP 4.x is supported")
	}
}

func setJobSuffix() {
	if CommandFlags.OcpVersion == "4.6" || CommandFlags.OcpVersion == "4.7" {
		CommandFlags.JobSuffix = "aws"
	} else {
		CommandFlags.JobSuffix = "claim"
	}
}

func jobPreRun(cmd *cobra.Command, args []string) {
	switch CommandFlags.CIEnvironment {
	case "common":
		CommandFlags.CIJobs = "common"
		CommandFlags.CIRepo = "preflight"
	case "preprod":
		CommandFlags.CIJobs = "preprod"
		CommandFlags.CIRepo = "certified-operators-preprod"
	case "prod":
		CommandFlags.CIJobs = "prod"
		CommandFlags.CIRepo = "certified-operators-prod"
	}

	validateJobFlags()

	config, err := GetGitHubFile("openshift", "release", "core-services/prow/02_config/_config.yaml")
	if err != nil {
		log.Fatalf("Error getting _config.yaml: %v", err)
	}

	if err = WriteToFileSystem(AppFs, config, "_config.yaml"); err != nil {
		log.Fatalf("Unable to write _config.yaml: %v", err)
	}

	periodic, err := GetGitHubFile("openshift", "release", "ci-operator/jobs/redhat-openshift-ecosystem/"+CommandFlags.CIRepo+"/redhat-openshift-ecosystem-"+CommandFlags.CIRepo+"-ocp-"+CommandFlags.OcpVersion+"-periodics.yaml")
	if err != nil {
		log.Fatalf("Error getting redhat-openshift-ecosystem-%s-ocp-%s-periodics.yaml: %v", CommandFlags.CIRepo, CommandFlags.OcpVersion, err)
	}

	if err = WriteToFileSystem(AppFs, periodic, "redhat-openshift-ecosystem-"+CommandFlags.CIRepo+"-ocp-"+CommandFlags.OcpVersion+"-periodics.yaml"); err != nil {
		log.Fatalf("Unable to write periodic job yaml: %v", err)
	}

	setJobSuffix()

	CommandFlags.JobName = "periodic-ci-redhat-openshift-ecosystem-" + CommandFlags.CIRepo + "-ocp-" + CommandFlags.OcpVersion + "-preflight-" + CommandFlags.CIJobs + "-" + CommandFlags.JobSuffix
	CommandFlags.JobConfigPath = "redhat-openshift-ecosystem-" + CommandFlags.CIRepo + "-ocp-" + CommandFlags.OcpVersion + "-periodics.yaml"
	CommandFlags.ConfigPath = "_config.yaml"
	if CommandFlags.OutputPath == "" {
		CommandFlags.OutputPath = "prowjob-base-url"
	}

	CommandFlags.ClusterType = CommandFlags.JobSuffix
}

func jobRun(cmd *cobra.Command, args []string) {
	configagent, err := CommandFlags.ConfigOptions.ConfigAgent()
	if err != nil {
		log.Fatalf("%v", err)
	}

	config := configagent.Config()
	jobmanifest, err := CreateProwJobManifest(CommandFlags.JobName, config)
	if err != nil {
		log.Fatalf("CreateProwJobManifest failed: %v", err)
	}

	if CommandFlags.PfltDockerConfig == "" {
		CommandFlags.PfltDockerConfig = os.Getenv("PFLT_DOCKERCONFIG")
	}

	multistageparams := map[string]string{
		"PFLT_LOGLEVEL":       CommandFlags.PfltLogLevel,
		"PFLT_LOGFILE":        CommandFlags.PfltLogFile,
		"PFLT_ARTIFACTS":      CommandFlags.PfltArtifacts,
		"PFLT_NAMESPACE":      CommandFlags.PfltNamespace,
		"PFLT_SERVICEACCOUNT": CommandFlags.PfltServiceAccount,
		"PFLT_INDEXIMAGE":     CommandFlags.PfltIndexImage,
		"PFLT_DOCKERCONFIG":   CommandFlags.PfltDockerConfig,
		"TEST_ASSET":          CommandFlags.TestAsset,
		"ASSET_TYPE":          CommandFlags.AssetType,
	}
	AppendMultiStageParams(jobmanifest.Spec.PodSpec, multistageparams)
	SetInputHash(jobmanifest.Spec.PodSpec, CommandFlags.ClusterType, CommandFlags.OcpVersion)

	if CommandFlags.DryRun {
		yamloutput, err := yaml.Marshal(jobmanifest)
		if err != nil {
			log.Printf("Failed marshalling yaml for --dry-run: %v", err)
		}
		log.Printf("%s", yamloutput)
		os.Exit(0)
	}

	clusterconfig, err := util.LoadClusterConfig()
	if err != nil {
		log.Fatalf("Error loading clusterconfig: %v", err)
	}

	pjcs, err := pjclient.NewForConfig(clusterconfig)
	if err != nil {
		log.Fatalf("Error creating prowjob client: %v", err)
	}

	pj, err := pjcs.ProwV1().ProwJobs(config.ProwJobNamespace).Create(context.Background(), jobmanifest, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("Error creating prowjob: %v", err)
	}

	selector := fields.SelectorFromSet(map[string]string{"metadata.name": pj.Name})

	var ok bool
	watcher, err := ProwJobWatcher(pj.Namespace, pjcs, selector.String())
	if err != nil {
		log.Fatalf("Error watching prowjob: %v", err)
	}

	log.Print("Waiting for prowjob status...")
	eventchannel := watcher.ResultChan()
	timeout := time.After(time.Duration(30) * time.Minute)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case event := <-eventchannel:
			pj, ok = event.Object.(*pjapi.ProwJob)
			if !ok {
				log.Fatalf("Received unexpected object type from watch: object-type %T", event.Object)
			}

			if pj.Status.State == pjapi.FailureState || pj.Status.State == pjapi.ErrorState || pj.Status.State == pjapi.AbortedState {
				ProwJobFailure(pj, config, CommandFlags.OutputPath)
			}

			if pj.Status.State == pjapi.SuccessState {
				ProwJobSuccess(pj, config, CommandFlags.OutputPath)
			}
		case <-timeout:
			ProwJobFailure(pj, config, CommandFlags.OutputPath)
		default:
			continue
		}
	}
}
