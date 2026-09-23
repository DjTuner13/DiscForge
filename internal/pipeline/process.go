package pipeline

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/DjTuner13/DiscForge/internal/jobs"
)

// RunPipe connects a producer such as vspipe to a consumer such as ffmpeg.
// Both commands share the caller's context, so cancellation stops the whole
// pipeline instead of leaving an orphaned child process behind.
func RunPipe(ctx context.Context, producer, consumer Command, logWriter io.Writer, consumerOutput io.Writer) error {
	return RunPipeWithProgress(ctx, producer, consumer, logWriter, consumerOutput, nil)
}

// RunPipeWithProgress is RunPipe with optional parsing of FFmpeg progress
// records emitted on stderr.
func RunPipeWithProgress(ctx context.Context, producer, consumer Command, logWriter io.Writer, consumerOutput io.Writer, progress func(jobs.Progress)) error {
	if producer.Name == "" || consumer.Name == "" {
		return fmt.Errorf("producer and consumer commands are required")
	}
	pipeReader, pipeWriter := io.Pipe()
	producerCmd := exec.CommandContext(ctx, producer.Name, producer.Args...)
	consumerCmd := exec.CommandContext(ctx, consumer.Name, consumer.Args...)
	producerCmd.Env = append(os.Environ(), producer.Env...)
	consumerCmd.Env = append(os.Environ(), consumer.Env...)
	producerCmd.Stdout = pipeWriter
	producerCmd.Stderr = logWriter
	consumerCmd.Stdin = pipeReader
	consumerCmd.Stdout = consumerOutput
	progressSink := &progressWriter{progress: progress}
	consumerCmd.Stderr = io.MultiWriter(nonNilWriter(logWriter), progressSink)
	if err := consumerCmd.Start(); err != nil {
		return fmt.Errorf("start consumer: %w", err)
	}
	if err := producerCmd.Start(); err != nil {
		_ = consumerCmd.Process.Kill()
		_ = consumerCmd.Wait()
		return fmt.Errorf("start producer: %w", err)
	}
	producerErr := producerCmd.Wait()
	_ = pipeWriter.CloseWithError(producerErr)
	consumerErr := consumerCmd.Wait()
	if producerErr != nil {
		return fmt.Errorf("producer: %w", producerErr)
	}
	if consumerErr != nil {
		return fmt.Errorf("consumer: %w", consumerErr)
	}
	return nil
}

func nonNilWriter(writer io.Writer) io.Writer {
	if writer == nil {
		return io.Discard
	}
	return writer
}

type progressWriter struct {
	progress func(jobs.Progress)
	buffer   string
	current  jobs.Progress
}

func (w *progressWriter) Write(p []byte) (int, error) {
	if w.progress == nil {
		return len(p), nil
	}
	w.buffer += string(p)
	for {
		line, rest, ok := strings.Cut(w.buffer, "\n")
		if !ok {
			break
		}
		w.buffer = rest
		if jobs.ParseProgressLine(line, &w.current) {
			w.progress(w.current)
		}
	}
	return len(p), nil
}
