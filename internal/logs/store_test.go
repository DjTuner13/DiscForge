package logs

import (
	"path/filepath"
	"testing"
)

func TestStoreAppendsPrivateJobLog(t *testing.T) {
	store := Store{Root: filepath.Join(t.TempDir(), "logs")}
	writer, err := store.Open("job-001")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("first\n")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	writer, err = store.Open("job-001")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = writer.Write([]byte("second\n"))
	_ = writer.Close()
	data, err := store.Read("job-001")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "first\nsecond\n" {
		t.Fatalf("log = %q", data)
	}
}
