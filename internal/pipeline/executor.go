package pipeline

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/DjTuner13/DiscForge/internal/jobs"
	"github.com/DjTuner13/DiscForge/internal/probe"
	"github.com/DjTuner13/DiscForge/internal/profiles"
)

type Inspector interface {
	Inspect(context.Context, string) (probe.Media, error)
}
type PipeFunc func(context.Context, Command, Command, io.Writer, io.Writer) error
type CommandFunc func(context.Context, Command, io.Writer) error

type RestorationExecutor struct {
	ArchiveRoot string
	Profile     profiles.Profile
	Script      string
	Inspector   Inspector
	RunPipe     PipeFunc
	RunCommand  CommandFunc
	Log         io.Writer
}

func (e RestorationExecutor) Execute(ctx context.Context, job jobs.Job) error {
	guard := GuardedRunner{ArchiveRoot: e.ArchiveRoot}
	if err := guard.ValidateOutput(job.InputPath, job.OutputPath); err != nil {
		return err
	}
	if e.Inspector == nil {
		e.Inspector = probe.Runner{}
	}
	source, err := e.Inspector.Inspect(ctx, job.InputPath)
	if err != nil {
		return err
	}
	commands := BuildRestorationCommands(job.InputPath, e.Script, TemporaryVideo(job.OutputPath), PartialOutput(job.OutputPath), e.Profile)
	if _, err := os.Stat(TemporaryVideo(job.OutputPath)); err == nil {
		return fmt.Errorf("temporary video already exists: %s", TemporaryVideo(job.OutputPath))
	} else if !os.IsNotExist(err) {
		return err
	}
	if _, err := os.Stat(PartialOutput(job.OutputPath)); err == nil {
		return fmt.Errorf("partial output already exists: %s", PartialOutput(job.OutputPath))
	} else if !os.IsNotExist(err) {
		return err
	}
	pipe := e.RunPipe
	if pipe == nil {
		pipe = RunPipe
	}
	if err := pipe(ctx, commands.VapourSynth, commands.Encode, e.Log, io.Discard); err != nil {
		return err
	}
	run := e.RunCommand
	if run == nil {
		run = guard.Run
	}
	if err := run(ctx, commands.Mux, e.Log); err != nil {
		return err
	}
	output, err := e.Inspector.Inspect(ctx, PartialOutput(job.OutputPath))
	if err != nil {
		return err
	}
	audio, subtitles := source.StreamCounts()
	rules := Validation{ExpectedVideoCodec: e.Profile.Encoder.Codec, ExpectedWidth: e.Profile.Video.Width, ExpectedHeight: e.Profile.Video.Height, ExpectedAudio: audio, ExpectedSubtitles: subtitles, ExpectedChapters: len(source.Chapters), ExpectedDuration: source.DurationSeconds(), DurationTolerance: 2}
	if err := ValidateOutput(PartialOutput(job.OutputPath), source, output, rules); err != nil {
		return err
	}
	if err := os.Rename(PartialOutput(job.OutputPath), job.OutputPath); err != nil {
		return fmt.Errorf("finalize output: %w", err)
	}
	return nil
}
