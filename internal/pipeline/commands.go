package pipeline

import (
	"fmt"
	"path/filepath"

	"github.com/DjTuner13/DiscForge/internal/profiles"
)

type RestorationCommands struct {
	VapourSynth Command
	Encode      Command
	Mux         Command
}

func BuildRestorationCommands(source, script, partialVideo, output string, profile profiles.Profile) RestorationCommands {
	vspipe := "vspipe"
	ffmpeg := "ffmpeg"
	return RestorationCommands{
		VapourSynth: Command{Name: vspipe, Args: []string{"-c", "y4m", script, "-"}},
		Encode:      Command{Name: ffmpeg, Args: []string{"-f", "yuv4mpegpipe", "-i", "-", "-an", "-c:v", profile.Encoder.Codec, "-preset", profile.Encoder.Preset, "-crf", fmt.Sprint(profile.Encoder.CRF), "-pix_fmt", profile.Video.PixelFormat, "-progress", "pipe:2", "-nostats", "-n", partialVideo}},
		Mux:         MuxCommand(partialVideo, source, output),
	}
}

func PartialOutput(output string) string { return output + ".partial" }
func TemporaryVideo(output string) string {
	return filepath.Join(filepath.Dir(output), "."+filepath.Base(output)+".video.partial")
}
