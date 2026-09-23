package pipeline

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// RunPipe connects a producer such as vspipe to a consumer such as ffmpeg.
// Both commands share the caller's context, so cancellation stops the whole
// pipeline instead of leaving an orphaned child process behind.
func RunPipe(ctx context.Context, producer, consumer Command, logWriter io.Writer, consumerOutput io.Writer) error {
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
	consumerCmd.Stderr = logWriter
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
