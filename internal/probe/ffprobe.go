package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Stream struct {
	Index      int    `json:"index"`
	CodecType  string `json:"codec_type"`
	CodecName  string `json:"codec_name"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	RFrameRate string `json:"r_frame_rate"`
	FieldOrder string `json:"field_order"`
}
type Chapter struct {
	ID        int               `json:"id"`
	StartTime string            `json:"start_time"`
	EndTime   string            `json:"end_time"`
	Tags      map[string]string `json:"tags"`
}
type Format struct {
	Duration   string `json:"duration"`
	Size       string `json:"size"`
	FormatName string `json:"format_name"`
}
type Media struct {
	Streams  []Stream  `json:"streams"`
	Chapters []Chapter `json:"chapters"`
	Format   Format    `json:"format"`
}
type Runner struct{ Binary string }

func (r Runner) Inspect(ctx context.Context, path string) (Media, error) {
	binary := r.Binary
	if binary == "" {
		binary = "ffprobe"
	}
	cmd := exec.CommandContext(ctx, binary, "-v", "error", "-print_format", "json", "-show_streams", "-show_chapters", "-show_format", path)
	out, err := cmd.Output()
	if err != nil {
		return Media{}, fmt.Errorf("ffprobe %s: %w", path, err)
	}
	var media Media
	if err := json.Unmarshal(out, &media); err != nil {
		return Media{}, fmt.Errorf("parse ffprobe output: %w", err)
	}
	return media, nil
}
func (m Media) StreamCounts() (audio, subtitles int) {
	for _, s := range m.Streams {
		if s.CodecType == "audio" {
			audio++
		}
		if s.CodecType == "subtitle" {
			subtitles++
		}
	}
	return
}
func (m Media) DurationSeconds() float64 {
	seconds, _ := strconv.ParseFloat(strings.TrimSpace(m.Format.Duration), 64)
	return seconds
}
