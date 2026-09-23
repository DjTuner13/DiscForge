package pipeline

import (
	"bytes"
	"context"
	"testing"
)

func TestRunPipeConnectsProcesses(t *testing.T) {
	var output, log bytes.Buffer
	err := RunPipe(context.Background(), Command{Name: "printf", Args: []string{"fixture"}}, Command{Name: "cat"}, &log, &output)
	if err != nil {
		t.Fatal(err)
	}
	if output.String() != "fixture" {
		t.Fatalf("output = %q", output.String())
	}
}

func TestRunPipePassesProducerEnvironment(t *testing.T) {
	var output bytes.Buffer
	err := RunPipe(context.Background(), Command{Name: "sh", Args: []string{"-c", "printf %s \"$DISCFORGE_SOURCE\""}, Env: []string{"DISCFORGE_SOURCE=fixture.mkv"}}, Command{Name: "cat"}, nil, &output)
	if err != nil {
		t.Fatal(err)
	}
	if output.String() != "fixture.mkv" {
		t.Fatalf("output = %q", output.String())
	}
}
