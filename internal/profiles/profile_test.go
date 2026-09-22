package profiles

import "testing"

func TestValidateRequiresStreamCopy(t *testing.T) {
	p := Profile{Name: "dvd"}
	p.Video.Width, p.Video.Height, p.Video.FPS = 720, 480, "60000/1001"
	p.VapourSynth.Script, p.Encoder.Codec, p.Encoder.Preset = "qtgmc.vpy", "libx265", "medium"
	p.Audio.Mode, p.Subtitles.Mode, p.Chapters.Mode = "copy", "copy", "copy"
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	p.Audio.Mode = "aac"
	if err := p.Validate(); err == nil {
		t.Fatal("expected non-copy audio to be rejected")
	}
}
