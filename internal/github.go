package internal

import (
	"context"

	"github.com/google/go-github/v56/github"
)

var client = github.NewClient(nil)

// GetGitHubFile accepts a repository owner and name and path and returns the contents of the file at that path.
func GetGitHubFile(owner, repo, path string) (string, error) {
	content, _, _, err := client.Repositories.GetContents(context.TODO(), owner, repo, path, nil)
	if err != nil {
		return "", err
	}

	data, err := content.GetContent()
	if err != nil {
		return "", err
	}

	return data, nil
}

// example of getting periodic job file from github
// curl -s https://api.github.com/repos/openshift/release/contents/ci-operator/jobs/redhat-openshift-ecosystem/preflight/redhat-openshift-ecosystem-preflight-ocp-4.10-periodics.yaml|jq -r '.content|split("\n")|join("")|@base64d'

// example of getting openshift-ci config file from github
// curl -s https://api.github.com/repos/openshift/release/contents/core-services/prow/02_config/_config.yaml|jq -r '.content|split("\n")|join("")|@base64d'
