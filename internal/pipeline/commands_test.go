package pipeline

import (
	"strings"
	"testing"

	"github.com/DjTuner13/DiscForge/internal/profiles"
)

func TestBuildRestorationCommandsUsesProfileAndPartialOutputs(t *testing.T) {
	profile := profiles.Profile{Name: "dvd"}
	profile.Video.PixelFormat = "yuv420p10le"
	profile.Encoder.Codec = "libx265"
	profile.Encoder.Preset = "medium"
	profile.Encoder.CRF = 16
	commands := BuildRestorationCommands("source.mkv", "qtgmc.vpy", ".video.partial", "final.mkv", profile)
	joined := strings.Join(commands.Encode.Args, " ")
	for _, want := range []string{"libx265", "medium", "16", "yuv420p10le", "-progress pipe:2", "-n"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("encode command %q missing %q", joined, want)
		}
	}
	if PartialOutput("final.mkv") != "final.mkv.partial" || TemporaryVideo("/work/final.mkv") != "/work/.final.mkv.video.partial" {
		t.Fatal("partial output naming changed")
	}
}
