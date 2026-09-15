package internal

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

type failingJobResult struct{}

func (failingJobResult) toJSON() ([]byte, error) {
	return nil, errors.New("marshal failed")
}

func TestWriteResultOutputReturnsMarshalError(t *testing.T) {
	err := writeResultOutput(failingJobResult{}, filepath.Join(t.TempDir(), "result.json"))
	if err == nil {
		t.Fatal("writeResultOutput returned nil for a marshal error")
	}
	if !strings.Contains(err.Error(), "unable to marshal prowjob result to JSON") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWriteResultOutputReturnsWriteError(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "missing", "result.json")
	err := writeResultOutput(&prowjobResult{}, outputPath)
	if err == nil {
		t.Fatal("writeResultOutput returned nil for a write error")
	}
	if !strings.Contains(err.Error(), outputPath) {
		t.Errorf("error does not contain the output path: %v", err)
	}
}
