package pipeline

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Command struct {
	Name string
	Args []string
}

// GuardedRunner centralizes process execution and refuses unsafe output paths.
type GuardedRunner struct{ ArchiveRoot string }

func (r GuardedRunner) ValidateOutput(input, output string) error {
	if output == "" {
		return fmt.Errorf("output path is required")
	}
	if filepath.Clean(input) == filepath.Clean(output) {
		return fmt.Errorf("output must differ from source")
	}
	archiveRoot, err := filepath.Abs(r.ArchiveRoot)
	if err != nil {
		return err
	}
	outputAbs, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	if outputAbs == archiveRoot || strings.HasPrefix(outputAbs, archiveRoot+string(os.PathSeparator)) {
		return fmt.Errorf("refusing to write beneath archive root %s", archiveRoot)
	}
	if _, err := os.Stat(output); err == nil {
		return fmt.Errorf("output already exists: %s", output)
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (r GuardedRunner) Run(ctx context.Context, command Command, logWriter io.Writer) error {
	if command.Name == "" {
		return fmt.Errorf("command name is required")
	}
	cmd := exec.CommandContext(ctx, command.Name, command.Args...)
	cmd.Stdout, cmd.Stderr = logWriter, logWriter
	return cmd.Run()
}

func (r GuardedRunner) RunSafe(ctx context.Context, input, output string, command Command, logWriter io.Writer) error {
	if err := r.ValidateOutput(input, output); err != nil {
		return err
	}
	return r.Run(ctx, command, logWriter)
}
