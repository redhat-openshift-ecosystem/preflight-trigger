package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownloadArtifactsLogsHTTPStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	temporaryDirectory := t.TempDir()
	if err := os.Chdir(temporaryDirectory); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(workingDirectory); err != nil {
			t.Errorf("unable to restore working directory: %v", err)
		}
	})

	var logs bytes.Buffer
	logWriter := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() {
		log.SetOutput(logWriter)
	})

	if downloadArtifacts(server.URL) {
		t.Fatal("downloadArtifacts returned success for a non-200 response")
	}
	if !strings.Contains(logs.String(), "404") {
		t.Errorf("log does not contain the HTTP status: %q", logs.String())
	}
	if strings.Contains(logs.String(), "<nil>") {
		t.Errorf("log contains a nil error instead of the HTTP status: %q", logs.String())
	}
}

func TestUntarArtifactsRejectsUnsupportedEntryType(t *testing.T) {
	temporaryDirectory := t.TempDir()
	tarballPath := filepath.Join(temporaryDirectory, "preflight.tar.gz")
	archive, err := os.Create(tarballPath)
	if err != nil {
		t.Fatal(err)
	}

	gzipWriter := gzip.NewWriter(archive)
	tarWriter := tar.NewWriter(gzipWriter)
	err = tarWriter.WriteHeader(&tar.Header{
		Name:     "unsupported",
		Typeflag: tar.TypeSymlink,
		Linkname: "target",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}

	if untarArtifacts(tarballPath, filepath.Join(temporaryDirectory, "output")) {
		t.Fatal("untarArtifacts returned success for an unsupported entry type")
	}
}
