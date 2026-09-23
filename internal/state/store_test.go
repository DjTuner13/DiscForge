package state

import (
	"path/filepath"
	"testing"

	"github.com/discforge/discforge/internal/jobs"
)

func TestStoreRecoversActiveJobsAsInterrupted(t *testing.T) {
	store := New(filepath.Join(t.TempDir(), "state.json"))
	want := Snapshot{Jobs: []jobs.Job{{ID: "one", Status: jobs.Running}, {ID: "two", Status: jobs.Queued}}}
	if err := store.Save(want); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Jobs[0].Status != jobs.Interrupted || got.Jobs[1].Status != jobs.Queued {
		t.Fatalf("recovered = %#v", got)
	}
}
