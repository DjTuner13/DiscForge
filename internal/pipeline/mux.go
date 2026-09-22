package pipeline

import "fmt"

// MuxCommand preserves source audio, subtitles, chapters, and metadata while
// taking the restored video stream from the first input.
func MuxCommand(restoredVideo, source, output string) Command {
	return Command{Name: "ffmpeg", Args: []string{"-i", restoredVideo, "-i", source, "-map", "0:v:0", "-map", "1:a?", "-map", "1:s?", "-map_metadata", "1", "-map_chapters", "1", "-c", "copy", "-n", output}}
}

func (c Command) String() string { return fmt.Sprintf("%s %v", c.Name, c.Args) }
