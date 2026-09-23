package main

import (
	"flag"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/discforge/discforge/internal/tui"
)

func main() {
	rootFlag := flag.String("archive-root", "", "read-only archive root")
	stateFlag := flag.String("state-file", "", "queue state file override")
	flag.Parse()
	root := *rootFlag
	if root == "" {
		root = os.Getenv("DISCFORGE_ARCHIVE_ROOT")
	}
	if root == "" {
		root = "/mnt/archive"
	}
	if *stateFlag != "" {
		_ = os.Setenv("DISCFORGE_STATE_FILE", *stateFlag)
	}

	p := tea.NewProgram(tui.NewModel(root), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
