package jobs

import "time"

type Status string

const (
	Queued     Status = "queued"
	Preparing  Status = "preparing"
	Running    Status = "running"
	Completed  Status = "completed"
	Failed     Status = "failed"
	Cancelled  Status = "cancelled"
	Interrupted Status = "interrupted"
)

type Job struct {
	ID          string
	InputPath   string
	OutputPath  string
	Profile     string
	Status      Status
	TotalFrames int64
	Frame       int64
	FPS         float64
	StartedAt   time.Time
	FinishedAt  time.Time
	Error       string
}

func (j Job) Percent() float64 {
	if j.TotalFrames == 0 {
		return 0
	}
	return float64(j.Frame) / float64(j.TotalFrames) * 100
}

func (j Job) Advance(now time.Time) Job {
	if j.Status == Queued {
		j.Status = Running
		j.StartedAt = now
	}
	if j.Status != Running {
		return j
	}
	j.Frame += 420
	j.FPS = 14.2
	if j.Frame >= j.TotalFrames {
		j.Frame = j.TotalFrames
		j.Status = Completed
		j.FinishedAt = now
	}
	return j
}

