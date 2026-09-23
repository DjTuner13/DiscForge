package pipeline

import (
	"fmt"
	"os"

	"github.com/DjTuner13/DiscForge/internal/probe"
)

type Validation struct {
	ExpectedVideoCodec string
	ExpectedWidth      int
	ExpectedHeight     int
	ExpectedAudio      int
	ExpectedSubtitles  int
	ExpectedChapters   int
	ExpectedDuration   float64
	DurationTolerance  float64
}

func ValidateOutput(path string, source, output probe.Media, rules Validation) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("output file: %w", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("output file is empty")
	}
	var video *probe.Stream
	for i := range output.Streams {
		if output.Streams[i].CodecType == "video" {
			video = &output.Streams[i]
			break
		}
	}
	if video == nil {
		return fmt.Errorf("output has no video stream")
	}
	if rules.ExpectedVideoCodec != "" && video.CodecName != rules.ExpectedVideoCodec {
		return fmt.Errorf("video codec %q, want %q", video.CodecName, rules.ExpectedVideoCodec)
	}
	if rules.ExpectedWidth > 0 && video.Width != rules.ExpectedWidth {
		return fmt.Errorf("video width %d, want %d", video.Width, rules.ExpectedWidth)
	}
	if rules.ExpectedHeight > 0 && video.Height != rules.ExpectedHeight {
		return fmt.Errorf("video height %d, want %d", video.Height, rules.ExpectedHeight)
	}
	audio, subs := output.StreamCounts()
	if audio != rules.ExpectedAudio {
		return fmt.Errorf("audio streams %d, want %d", audio, rules.ExpectedAudio)
	}
	if subs != rules.ExpectedSubtitles {
		return fmt.Errorf("subtitle streams %d, want %d", subs, rules.ExpectedSubtitles)
	}
	if rules.ExpectedChapters > 0 && len(output.Chapters) != rules.ExpectedChapters {
		return fmt.Errorf("chapters %d, want %d", len(output.Chapters), rules.ExpectedChapters)
	}
	if rules.ExpectedDuration > 0 && abs(output.DurationSeconds()-rules.ExpectedDuration) > rules.DurationTolerance {
		return fmt.Errorf("duration %.3fs is outside tolerance", output.DurationSeconds())
	}
	_ = source
	return nil
}
func abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
