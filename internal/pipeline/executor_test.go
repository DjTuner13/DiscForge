package pipeline

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/DjTuner13/DiscForge/internal/jobs"
	"github.com/DjTuner13/DiscForge/internal/probe"
	"github.com/DjTuner13/DiscForge/internal/profiles"
)

type fakeInspector struct{ media map[string]probe.Media }

func (f fakeInspector) Inspect(_ context.Context, path string) (probe.Media, error) {
	return f.media[path], nil
}
func testProfile() profiles.Profile {
	p := profiles.Profile{Name: "fixture"}
	p.Video.Width, p.Video.Height, p.Video.FPS = 320, 240, "30/1"
	p.VapourSynth.Script = "fixture.vpy"
	p.Encoder.Codec, p.Encoder.Preset, p.Encoder.CRF = "h264", "fast", 18
	p.Audio.Mode, p.Subtitles.Mode, p.Chapters.Mode = "copy", "copy", "copy"
	return p
}

func TestRestorationExecutorFinalizesOnlyAfterValidation(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "source.mkv")
	outputPath := filepath.Join(dir, "output.mkv")
	if err := os.WriteFile(sourcePath, []byte("source"), 0o600); err != nil {
		t.Fatal(err)
	}
	source := probe.Media{Streams: []probe.Stream{{CodecType: "video"}, {CodecType: "audio"}}, Format: probe.Format{Duration: "2"}}
	output := probe.Media{Streams: []probe.Stream{{CodecType: "video", CodecName: "h264", Width: 320, Height: 240}, {CodecType: "audio"}}, Format: probe.Format{Duration: "2"}}
	partial := PartialOutput(outputPath)
	executor := RestorationExecutor{ArchiveRoot: filepath.Join(dir, "archive"), Profile: testProfile(), Inspector: fakeInspector{media: map[string]probe.Media{sourcePath: source, partial: output}}, RunPipe: func(_ context.Context, _, _ Command, _ io.Writer, _ io.Writer) error { return nil }, RunCommand: func(_ context.Context, _ Command, _ io.Writer) error {
		return os.WriteFile(partial, []byte("finished"), 0o600)
	}}
	if err := executor.Execute(context.Background(), jobs.Job{InputPath: sourcePath, OutputPath: outputPath}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(partial); !os.IsNotExist(err) {
		t.Fatalf("partial output remains: %v", err)
	}
}

func TestCleanupArtifactsLeavesFinalOutputAlone(t *testing.T) {
	dir := t.TempDir()
	job := jobs.Job{OutputPath: filepath.Join(dir, "output.mkv")}
	if err := os.WriteFile(TemporaryVideo(job.OutputPath), []byte("video"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(PartialOutput(job.OutputPath), []byte("partial"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(job.OutputPath, []byte("final"), 0o600); err != nil {
		t.Fatal(err)
	}
	CleanupArtifacts(job)
	if _, err := os.Stat(TemporaryVideo(job.OutputPath)); !os.IsNotExist(err) {
		t.Fatalf("temporary output remains: %v", err)
	}
	if _, err := os.Stat(PartialOutput(job.OutputPath)); !os.IsNotExist(err) {
		t.Fatalf("partial output remains: %v", err)
	}
	if _, err := os.Stat(job.OutputPath); err != nil {
		t.Fatalf("final output was removed: %v", err)
	}
}
