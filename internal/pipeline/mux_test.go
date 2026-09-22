package pipeline

import "testing"

func TestMuxCommandPreservesSourceStreams(t *testing.T) {
	cmd := MuxCommand("restored.mkv", "source.mkv", "final.mkv")
	for _, want := range []string{"-map", "1:a?", "1:s?", "-map_chapters", "-n"} {
		found := false
		for _, arg := range cmd.Args {
			if arg == want {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing %q in %v", want, cmd.Args)
		}
	}
}
