package jobs

import "time"

func EstimateETA(totalFrames, frame int64, fps float64) time.Duration {
	if totalFrames <= frame || fps <= 0 {
		return 0
	}
	return time.Duration(float64(totalFrames-frame) / fps * float64(time.Second))
}
