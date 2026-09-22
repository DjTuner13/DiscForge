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
