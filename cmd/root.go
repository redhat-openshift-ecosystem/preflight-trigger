/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"
	"reflect"

	"github.com/spf13/cobra"
	configflagutil "sigs.k8s.io/prow/pkg/flagutil/config"

	"github.com/redhat-openshift-ecosystem/preflight-trigger/version"
)

type FlagsData struct {
	AssetType   string `json:"asset-type"`
	ClusterType string `json:"cluster-type"`
	// RootFlags inherits configflagutil.ConfigOptions from Prow and provides the following flags:
	// ConfigPath (string), JobConfigPath (string), ConfigPathFlagName (string), JobConfigPathFlagName (string),
	// SupplementalProwConfigDirs (flagutil.Strings), and SupplementalProwConfigsFileNameSuffix (string)
	configflagutil.ConfigOptions
	DocsType                string   `json:"docs-type" param:"DOCS_TYPE"`
	DryRun                  bool     `json:"dry-run" param:"DRY_RUN"`
	CIEnvironment           string   `json:"ci-environment" param:"CI_ENVIRONMENT"`
	CIJobs                  string   `json:"ci-jobs" param:"CI_JOBS"`
	CIRepo                  string   `json:"ci-repo" param:"CI_REPO"`
	FileToEncode            string   `json:"file-to-encode" param:"FILE_TO_ENCODE"`
	FileToEncrypt           string   `json:"file-to-encrypt" param:"FILE_TO_ENCRYPT"`
	FileToDecode            string   `json:"file-to-decode" param:"FILE_TO_DECODE"`
	FileToDecrypt           string   `json:"file-to-decrypt" param:"FILE_TO_DECRYPT"`
	GPGEncryptionPublicKey  string   `json:"gpg-encryption-public-key" param:"GPG_ENCRYPTION_PUBLIC_KEY"`
	GPGEncryptionPrivateKey string   `json:"gpg-encryption-private-key" param:"GPG_ENCRYPTION_PRIVATE_KEY"`
	GPGDecryptionPublicKey  string   `json:"gpg-decryption-public-key" param:"GPG_DECRYPTION_PUBLIC_KEY"`
	GPGDecryptionPrivateKey string   `json:"gpg-decryption-private-key" param:"GPG_DECRYPTION_PRIVATE_KEY"`
	GPGPassphrase           string   `json:"gpg-passphrase" param:"GPG_PASSPHRASE"`
	Hidden                  bool     `json:"hidden" param:"HIDDEN"`
	JobName                 string   `json:"job-name" param:"JOB_NAME"`
	JobNames                []string `json:"job-names" param:"JOB_NAMES"`
	JobSuffix               string   `json:"job-suffix" param:"JOB_SUFFIX"`
	OcpVersion              string   `json:"ocp-version" param:"OCP_VERSION"`
	OutputPath              string   `json:"output-path" param:"OUTPUT_PATH"`
	PfltArtifacts           string   `json:"pflt-artifacts" param:"PFLT_ARTIFACTS"`
	PfltDockerConfig        string   `json:"pflt-docker-config" param:"PFLT_DOCKERCONFIG"`
	PfltIndexImage          string   `json:"pflt-index-image" param:"PFLT_INDEX_IMAGE"`
	PfltLogFile             string   `json:"pflt-log-file" param:"PFLT_LOG_FILE"`
	PfltLogLevel            string   `json:"pflt-log-level" param:"PFLT_LOG_LEVEL"`
	PfltNamespace           string   `json:"pflt-namespace" param:"PFLT_NAMESPACE"`
	PfltServiceAccount      string   `json:"pflt-service-account" param:"PFLT_SERVICE_ACCOUNT"`
	ReleaseImageRef         string   `json:"release-image-ref" param:"RELEASE_IMAGE_REF"`
	TestAsset               string   `json:"test-asset" param:"TEST_ASSET"`
	UseStdin                bool     `json:"use-stdin" param:"USE_STDIN"`
	UseStdout               bool     `json:"use-stdout" param:"USE_STDOUT"`
	ValueToEncode           string   `json:"value-to-encode" param:"VALUE_TO_ENCODE"`
	ValueToDecode           string   `json:"value-to-decode" param:"VALUE_TO_DECODE"`
}

var CommandFlags FlagsData

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "preflight-trigger",
	Short:   "Create on-demand preflight jobs in openshift-ci system",
	Long:    ``,
	Version: version.Version.String(),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.DisableAutoGenTag = true
	flags := rootCmd.PersistentFlags()
	flags.StringVarP(&CommandFlags.AssetType, "asset-type", "", "", "Type of asset to trigger")
	flags.BoolVarP(&CommandFlags.DryRun, "dry-run", "", false, "Do perform any actions, but do not actually trigger the job")
	flags.StringVarP(&CommandFlags.GPGEncryptionPublicKey, "gpg-encryption-public-key", "", "", "GPG public key to use for encryption")
	flags.StringVarP(&CommandFlags.GPGEncryptionPrivateKey, "gpg-encryption-private-key", "", "", "GPG private key to use for encryption")
	flags.StringVarP(&CommandFlags.GPGDecryptionPublicKey, "gpg-decryption-public-key", "", "", "GPG public key to use for decryption")
	flags.StringVarP(&CommandFlags.GPGDecryptionPrivateKey, "gpg-decryption-private-key", "", "", "GPG private key to use for decryption")
	flags.BoolVarP(&CommandFlags.Hidden, "hidden", "", false, "Hide job in the list of jobs visible by deck")
	flags.StringVarP(&CommandFlags.JobName, "job-name", "", "", "Name of the job to trigger")
	flags.StringVarP(&CommandFlags.JobSuffix, "job-suffix", "", "", "Suffix to append to the job name")
	flags.StringVarP(&CommandFlags.OcpVersion, "ocp-version", "", "", "Version of OCP to use")
	flags.StringVarP(&CommandFlags.OutputPath, "output-path", "", "", "Path to output the job to")
	flags.StringVarP(&CommandFlags.PfltArtifacts, "pflt-artifacts", "", "artifacts", "Path to artifacts to use for preflight")
	flags.StringVarP(&CommandFlags.PfltDockerConfig, "pflt-docker-config", "", "", "Docker config to use for preflight")
	flags.StringVarP(&CommandFlags.PfltIndexImage, "pflt-index-image", "", "", "Index image to use for preflight")
	flags.StringVarP(&CommandFlags.PfltLogFile, "pflt-log-file", "", "", "Path to log file to use for preflight")
	flags.StringVarP(&CommandFlags.PfltLogLevel, "pflt-log-level", "", "trace", "Level of logging to use for preflight")
	flags.StringVarP(&CommandFlags.PfltNamespace, "pflt-namespace", "", "", "Namespace to use for preflight")
	flags.StringVarP(&CommandFlags.PfltServiceAccount, "pflt-service-account", "", "", "Service account to use for preflight")
	flags.StringVarP(&CommandFlags.ReleaseImageRef, "release-image-ref", "", "", "Release image reference to use for preflight")
	flags.StringVarP(&CommandFlags.TestAsset, "test-asset", "", "", "Test asset to use for preflight")
}

func (f FlagsData) Get(flag string) reflect.Value {
	s := reflect.ValueOf(f)
	value := s.FieldByName(flag)
	return value
}

func getStdin() string {
	data, err := os.ReadFile(os.Stdin.Name())
	if err != nil {
		log.Fatalf("Unable to read stdin: %v", err)
	}
	return string(data)
}
