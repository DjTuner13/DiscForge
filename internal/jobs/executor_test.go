package jobs

import (
	"context"
	"errors"
	"testing"
	"time"
)

type testExecutor struct{ err error }

func (e testExecutor) Execute(context.Context, Job) error { return e.err }

func TestRunOneRecordsFailure(t *testing.T) {
	q := Queue{}
	q.Add(Job{ID: "job-1"})
	expected := errors.New("validation failed")
	if err := q.RunOne(context.Background(), testExecutor{err: expected}, time.Unix(1, 0)); !errors.Is(err, expected) {
		t.Fatalf("error = %v", err)
	}
	if q.Jobs[0].Status != Failed || q.Jobs[0].Error != expected.Error() {
		t.Fatalf("job = %#v", q.Jobs[0])
	}
}
