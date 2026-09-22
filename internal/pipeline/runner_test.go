package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGuardedRunnerRejectsArchiveWritesAndOverwrite(t *testing.T) {
	root := t.TempDir()
	runner := GuardedRunner{ArchiveRoot: root}
	if err := runner.ValidateOutput(filepath.Join(root, "master.mkv"), filepath.Join(root, "out.mkv")); err == nil {
		t.Fatal("expected archive write to be rejected")
	}
	existing := filepath.Join(t.TempDir(), "out.mkv")
	f, err := os.Create(existing)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if err := runner.ValidateOutput("source.mkv", existing); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("overwrite error = %v", err)
	}
}
