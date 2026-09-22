package main

import (
	"log"
	"os"

	"github.com/djranoia/discforge/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	root := os.Getenv("DISCFORGE_ARCHIVE_ROOT")
	if root == "" {
		root = "/mnt/archive"
	}

	p := tea.NewProgram(tui.NewModel(root), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

