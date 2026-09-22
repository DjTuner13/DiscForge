package jobs

import (
	"testing"
	"time"
)

func TestAdvanceCompletesJob(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	j := Job{Status: Queued, TotalFrames: 1000}
	j = j.Advance(now)
	if j.Status != Running || j.Frame != 420 {
		t.Fatalf("first advance = %#v", j)
	}
	j.Frame = 900
	j = j.Advance(now.Add(time.Minute))
	if j.Status != Completed || j.Frame != 1000 || j.FinishedAt.IsZero() {
		t.Fatalf("completion = %#v", j)
	}
}
