package jobs

import "time"

type Queue struct{ Jobs []Job }

func (q *Queue) Add(job Job) { job.Status = Queued; q.Jobs = append(q.Jobs, job) }

// Advance runs at most one job, matching the MVP's single-encode policy.
func (q *Queue) Advance(now time.Time) {
	for i := range q.Jobs {
		if q.Jobs[i].Status == Running {
			q.Jobs[i] = q.Jobs[i].Advance(now)
			return
		}
	}
	for i := range q.Jobs {
		if q.Jobs[i].Status == Queued {
			q.Jobs[i] = q.Jobs[i].Advance(now)
			return
		}
	}
}

func (q Queue) Active() *Job {
	for i := range q.Jobs {
		if q.Jobs[i].Status == Running {
			return &q.Jobs[i]
		}
	}
	return nil
}
