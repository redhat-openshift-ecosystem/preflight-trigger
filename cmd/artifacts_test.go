package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
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

// TestUntarArtifactsRejectsPathTraversal verifies archive paths cannot escape the target.
func TestUntarArtifactsRejectsPathTraversal(t *testing.T) {
	for _, test := range []struct {
		name     string
		typeflag byte
	}{
		{name: "regular file", typeflag: tar.TypeReg},
		{name: "directory", typeflag: tar.TypeDir},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			target := filepath.Join(root, "nested", "output")
			if err := os.MkdirAll(target, 0o755); err != nil {
				t.Fatal(err)
			}

			tarball := filepath.Join(root, "archive.tar.gz")
			header := &tar.Header{
				Name:     "../../outside",
				Typeflag: test.typeflag,
			}
			if test.typeflag == tar.TypeReg {
				header.Size = int64(len("data"))
			}
			writeTarGz(t, tarball, header)

			if untarArtifacts(tarball, target) {
				t.Fatal("untarArtifacts returned success for a path traversal entry")
			}

			if _, err := os.Stat(filepath.Join(root, "outside")); !os.IsNotExist(err) {
				t.Errorf("path traversal created a file or directory outside the target: %v", err)
			}
		})
	}
}

// TestUntarArtifactsExtractsWithEmptyTarget verifies an empty target uses the current directory.
func TestUntarArtifactsExtractsWithEmptyTarget(t *testing.T) {
	root := t.TempDir()
	tarball := filepath.Join(root, "archive.tar.gz")
	writeTarGz(t, tarball, &tar.Header{
		Name:     "inside",
		Typeflag: tar.TypeReg,
		Size:     int64(len("data")),
	})

	t.Chdir(root)
	if !untarArtifacts(tarball, "") {
		t.Fatal("untarArtifacts rejected a safe entry with an empty target")
	}

	contents, err := os.ReadFile(filepath.Join(root, "inside"))
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "data" {
		t.Fatalf("unexpected extracted contents: %q", contents)
	}
}

// TestUntarArtifactsExtractsDirectoryEntry verifies trailing separators are handled.
func TestUntarArtifactsExtractsDirectoryEntry(t *testing.T) {
	root := t.TempDir()
	tarball := filepath.Join(root, "archive.tar.gz")
	writeTarGz(t, tarball, &tar.Header{
		Name:     "artifacts/",
		Typeflag: tar.TypeDir,
	})

	t.Chdir(root)
	if !untarArtifacts(tarball, "") {
		t.Fatal("untarArtifacts rejected a directory entry with a trailing separator")
	}
	if info, err := os.Stat(filepath.Join(root, "artifacts")); err != nil {
		t.Fatal(err)
	} else if !info.IsDir() {
		t.Fatal("extracted directory entry is not a directory")
	}
}

// TestUntarArtifactsRejectsSymlinkEscape verifies extraction cannot follow a symlink outside the target.
func TestUntarArtifactsRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	outside := filepath.Join(root, "outside")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(target, "link")); err != nil {
		t.Skipf("unable to create symlink: %v", err)
	}

	tarball := filepath.Join(root, "archive.tar.gz")
	writeTarGz(t, tarball, &tar.Header{
		Name:     "link/file",
		Typeflag: tar.TypeReg,
		Size:     int64(len("data")),
	})

	if untarArtifacts(tarball, target) {
		t.Fatal("untarArtifacts returned success for a symlink escape")
	}
	if _, err := os.Stat(filepath.Join(outside, "file")); !os.IsNotExist(err) {
		t.Errorf("symlink escape created a file outside the target: %v", err)
	}
}

func TestBuildArtifactsURL(t *testing.T) {
	const artifactPath = "periodic-ci-redhat-openshift-ecosystem-preflight-ocp-4.18-preflight-common-claim/job-123/artifacts/preflight-common-claim/operator-pipelines-preflight-common-encrypt/artifacts/preflight.tar.gz.asc"

	tests := []struct {
		name    string
		baseURL string
		want    string
	}{
		{
			name:    "default base URL",
			baseURL: artifactsBaseURL,
			want:    artifactsBaseURL + artifactPath,
		},
		{
			name:    "custom base URL",
			baseURL: "https://example.test/results",
			want:    "https://example.test/results/" + artifactPath,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flags := &FlagsData{
				ArtifactsBaseURL: test.baseURL,
				CIRepo:           "preflight",
				OcpVersion:       "4.18",
				CIJobs:           "common",
				JobSuffix:        "claim",
			}

			got, err := buildArtifactsURL(flags, "job-123")
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Errorf("buildArtifactsURL() = %q, want %q", got, test.want)
			}
		})
	}
}

// writeTarGz creates a gzip-compressed tar archive containing one entry.
func writeTarGz(t *testing.T, path string, header *tar.Header) {
	t.Helper()

	archive, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	gzipWriter := gzip.NewWriter(archive)
	tarWriter := tar.NewWriter(gzipWriter)

	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatal(err)
	}
	if header.Typeflag == tar.TypeReg {
		if _, err := io.WriteString(tarWriter, "data"); err != nil {
			t.Fatal(err)
		}
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
}
