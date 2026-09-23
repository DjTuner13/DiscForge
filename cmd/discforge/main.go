package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/DjTuner13/DiscForge/internal/archive"
	"github.com/DjTuner13/DiscForge/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	rootFlag := flag.String("archive-root", "", "read-only archive root")
	stateFlag := flag.String("state-file", "", "queue state file override")
	scanFlag := flag.Bool("scan", false, "list MKV archive masters and exit")
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
	if *scanFlag {
		files, err := archive.Scan(root)
		if err != nil {
			log.Fatal(err)
		}
		for _, path := range files {
			fmt.Println(path)
		}
		fmt.Fprintf(os.Stderr, "Found %d MKV archive master(s) under %s\n", len(files), root)
		return
	}

	p := tea.NewProgram(tui.NewModel(root), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
