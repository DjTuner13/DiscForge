package jobs

import (
	"context"
	"fmt"
	"time"
)

type Executor interface {
	Execute(context.Context, Job) error
}

// RunOne enforces the single-job policy and records a terminal failure rather
// than incorrectly marking a process complete when execution fails.
func (q *Queue) RunOne(ctx context.Context, executor Executor, now time.Time) error {
	for i := range q.Jobs {
		if q.Jobs[i].Status == Running {
			return fmt.Errorf("job %s is already running", q.Jobs[i].ID)
		}
	}
	index := -1
	for i := range q.Jobs {
		if q.Jobs[i].Status == Queued {
			index = i
			break
		}
	}
	if index < 0 {
		return nil
	}
	q.Jobs[index].Status = Preparing
	q.Jobs[index].Status = Running
	if err := executor.Execute(ctx, q.Jobs[index]); err != nil {
		q.Jobs[index].Status = Failed
		q.Jobs[index].Error = err.Error()
		return err
	}
	q.Jobs[index].Status = Completed
	q.Jobs[index].FinishedAt = now
	return nil
}
