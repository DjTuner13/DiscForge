package jobs

import (
	"bufio"
	"strconv"
	"strings"
	"time"
)

// Progress is the machine-readable subset emitted by ffmpeg -progress.
type Progress struct {
	Frame   int64
	FPS     float64
	OutTime time.Duration
	Speed   float64
	Done    bool
}

func ParseProgressLine(line string, progress *Progress) bool {
	key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
	if !ok {
		return false
	}
	switch key {
	case "frame":
		progress.Frame, _ = strconv.ParseInt(value, 10, 64)
	case "fps":
		progress.FPS, _ = strconv.ParseFloat(value, 64)
	case "out_time_us":
		if micros, err := strconv.ParseInt(value, 10, 64); err == nil {
			progress.OutTime = time.Duration(micros) * time.Microsecond
		}
	case "speed":
		progress.Speed, _ = strconv.ParseFloat(strings.TrimSuffix(value, "x"), 64)
	case "progress":
		progress.Done = value == "end"
	default:
		return false
	}
	return true
}

func ParseProgress(r *bufio.Scanner, progress *Progress) {
	for r.Scan() {
		ParseProgressLine(r.Text(), progress)
	}
}
