package testutil

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAbsFilePath(t *testing.T) {
	result := AbsFilePath(t, ".")
	if result == "" {
		t.Fatal("expected non-empty absolute path")
	}
	if !filepath.IsAbs(result) {
		t.Fatalf("expected absolute path, got %s", result)
	}
}

func TestAbsFilePath_ParentDir(t *testing.T) {
	result := AbsFilePath(t, "..")
	if !strings.Contains(result, filepath.Clean("..")) {
		// Just ensure it resolves without panic
	}
	if !filepath.IsAbs(result) {
		t.Fatalf("expected absolute path, got %s", result)
	}
}
