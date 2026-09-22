package archive

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFindsMKVRecursivelyAndIgnoresOtherFiles(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "Show", "Season 01")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"A2.MKV", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(nested, name), []byte("fixture"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	files, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || filepath.Base(files[0]) != "A2.MKV" {
		t.Fatalf("files = %#v", files)
	}
}
