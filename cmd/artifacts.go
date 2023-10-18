/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	pjapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
)

// artifactsCmd represents the artifacts command
var artifactsCmd = &cobra.Command{
	Use:   "artifacts",
	Short: "Get artifacts from a given openshift-ci job",
	Long: `Get artifacts from a given openshift-ci job. This command
is used to get the artifacts from a given job and download them to the
local filesystem. This command also supports untarring the artifacts.`,
	PreRunE: artifactsPreRunE,
	RunE:    artifactsRunE,
}

func init() {
	rootCmd.AddCommand(artifactsCmd)
	artifactsCmd.Flags().BoolP("untar", "", false, "Untar the artifacts file")
	artifactsCmd.Flags().StringVarP(&CommandFlags.CIEnvironment, "environment", "", "common", "Set the environment to use; can be one of [common, preprod, prod]")
}

func getJobID() string {
	var results struct {
		ArtifactsURL string             `json:"prowjob_artifacts_url"`
		Status       pjapi.ProwJobState `json:"status"`
		URL          string             `json:"prowjob_url"`
	}

	f, err := os.ReadFile("prowjob-base-url")
	if err != nil {
		log.Fatalf("Unable to read prowjob-base-url file: %v", err)
	}

	err = json.Unmarshal(f, &results)
	if err != nil {
		log.Fatalf("Unable to unmarshal prowjob-base-url file: %v", err)
	}

	return func(sl []string) string {
		return sl[len(sl)-1]
	}(strings.Split(results.ArtifactsURL, "/"))
}

func downloadArtifacts(uri string) bool {
	out, err := os.Create("preflight.tar.gz.asc")
	if err != nil {
		log.Printf("Unable to create preflight.tar.gz.asc file: %v", err)
		return false
	}

	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			log.Fatalf("Unable to close preflight.tar.gz.asc file: %v", err)
		}
	}(out)

	resp, err := http.Get(uri)
	if err != nil {
		log.Printf("Unable to download preflight.tar.gz.asc file: %v", err)
		return false
	}

	defer func(resp *http.Response) {
		err := resp.Body.Close()
		if err != nil {
			log.Fatalf("Unable to close preflight.tar.gz.asc file: %v", err)
		}
	}(resp)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Response returned status code other than 200: %v", err)
		return false
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Printf("Unable to copy data to preflight.tar.gz.asc file: %v", err)
		return false
	}

	return true
}

func untarArtifacts(tarball, target string) bool {
	src := filepath.FromSlash(tarball)
	archive, err := os.Open(src)
	if err != nil {
		log.Printf("Unable to open preflight.tar.gz file: %v", err)
		return false
	}

	defer func(a *os.File) {
		err := a.Close()
		if err != nil {
			log.Fatalf("Unable to close preflight.tar.gz file: %v", err)
		}
	}(archive)

	gzreader, err := gzip.NewReader(archive)
	if err != nil {
		log.Printf("Unable to create gzip reader: %v", err)
		return false
	}

	defer func(g *gzip.Reader) {
		err := g.Close()
		if err != nil {
			log.Fatalf("Unable to close gzip reader: %v", err)
		}
	}(gzreader)

	tarreader := tar.NewReader(gzreader)

	for {
		header, err := tarreader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Unable to read tar file: %v", err)
			return false
		}

		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(filepath.Join(target, header.Name), 0o755)
			if err != nil {
				log.Printf("Unable to create directory: %v", err)
				return false
			}
		case tar.TypeReg:
			outFile, err := os.Create(filepath.Join(target, header.Name))
			if err != nil {
				log.Printf("Unable to create file: %v", err)
				return false
			}
			_, err = io.Copy(outFile, tarreader)
			if err != nil {
				log.Printf("Unable to copy file: %v", err)
				return false
			}
			err = outFile.Close()
			if err != nil {
				log.Fatalf("Unable to close file: %v", err)
			}
		default:
			log.Printf("Unable to untar type: %v in file %v", header.Typeflag, header.Name)
		}
	}
	return true
}

func artifactsPreRunE(cmd *cobra.Command, args []string) error {
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

	setJobSuffix()
	CommandFlags.ClusterType = CommandFlags.JobSuffix

	return nil
}

func artifactsRunE(cmd *cobra.Command, args []string) error {
	ok, err := cmd.Flags().GetBool("untar")
	if err != nil {
		log.Fatalf("Unable to get untar flag: %v", err)
	}

	if !ok {
		artifactsBaseURL := "https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/logs/"
		artifactsJobID := getJobID()
		artifactsTarballURI := artifactsBaseURL + "periodic-ci-redhat-openshift-ecosystem-" + CommandFlags.CIRepo +
			"-ocp-" + CommandFlags.OcpVersion + "-preflight-" + CommandFlags.CIJobs + "-" + CommandFlags.JobSuffix + "/" + artifactsJobID +
			"/artifacts/preflight-" + CommandFlags.CIJobs + "-" + CommandFlags.JobSuffix + "/operator-pipelines-preflight-" + CommandFlags.CIJobs + "-encrypt/artifacts/preflight.tar.gz.asc"

		dok := downloadArtifacts(artifactsTarballURI)
		if !dok {
			log.Fatalf("Unable to download preflight.tar.gz.asc file")
		}
	} else {
		uok := untarArtifacts("preflight.tar.gz", "")
		if !uok {
			log.Fatalf("Unable to untar preflight.tar.gz file")
		}
	}

	return nil
}
