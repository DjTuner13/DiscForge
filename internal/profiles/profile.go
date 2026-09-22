package profiles

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	Name  string `yaml:"name"`
	Video struct {
		Width       int    `yaml:"width"`
		Height      int    `yaml:"height"`
		FPS         string `yaml:"fps"`
		PixelFormat string `yaml:"pixel_format"`
	} `yaml:"video"`
	Source struct {
		FieldOrder string `yaml:"field_order"`
	} `yaml:"source"`
	VapourSynth struct {
		Script string `yaml:"script"`
	} `yaml:"vapoursynth"`
	Encoder struct {
		Codec  string `yaml:"codec"`
		Preset string `yaml:"preset"`
		CRF    int    `yaml:"crf"`
	} `yaml:"encoder"`
	Audio struct {
		Mode string `yaml:"mode"`
	} `yaml:"audio"`
	Subtitles struct {
		Mode string `yaml:"mode"`
	} `yaml:"subtitles"`
	Chapters struct {
		Mode string `yaml:"mode"`
	} `yaml:"chapters"`
}

func Load(path string) (Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Profile{}, err
	}
	var profile Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return Profile{}, err
	}
	if err := profile.Validate(); err != nil {
		return Profile{}, fmt.Errorf("%s: %w", path, err)
	}
	return profile, nil
}

func (p Profile) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}
	if p.Video.Width <= 0 || p.Video.Height <= 0 {
		return fmt.Errorf("video dimensions must be positive")
	}
	if p.Video.FPS == "" || p.VapourSynth.Script == "" {
		return fmt.Errorf("video fps and VapourSynth script are required")
	}
	if p.Encoder.Codec == "" || p.Encoder.Preset == "" {
		return fmt.Errorf("encoder codec and preset are required")
	}
	if p.Audio.Mode != "copy" || p.Subtitles.Mode != "copy" || p.Chapters.Mode != "copy" {
		return fmt.Errorf("audio, subtitles, and chapters must use copy mode")
	}
	return nil
}
