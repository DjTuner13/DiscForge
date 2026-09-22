package jobs

import (
	"testing"
	"time"
)

func TestQueueAdvancesOnlyOneJob(t *testing.T) {
	q := Queue{}
	q.Add(Job{ID: "one", TotalFrames: 1000})
	q.Add(Job{ID: "two", TotalFrames: 1000})
	q.Advance(time.Unix(1, 0))
	if q.Jobs[0].Status != Running || q.Jobs[1].Status != Queued {
		t.Fatalf("queue = %#v", q.Jobs)
	}
	q.Advance(time.Unix(2, 0))
	if q.Jobs[0].Frame != 840 || q.Jobs[1].Status != Queued {
		t.Fatalf("queue = %#v", q.Jobs)
	}
}
