package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Show usage
	if len(os.Args) >= 2 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		printUsage()
		return
	}

	// Role 2: annotation panel TUI (invoked internally by tmux split-pane)
	if len(os.Args) >= 3 && os.Args[1] == "--internal-watch" {
		paneID := os.Args[2]
		runAnnotationTUI(paneID)
		return
	}

	// Role 1: launcher
	runLauncher()
}

func runAnnotationTUI(paneID string) {
	m := NewWatchModel(paneID)

	opts := []tea.ProgramOption{tea.WithAltScreen()}
	p := tea.NewProgram(m, opts...)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}
}

func runLauncher() {
	clis := detectCLIs()
	if len(clis) == 0 {
		fmt.Fprintln(os.Stderr, "No installed AI CLI found (claude, gemini, codex, aider)")
		os.Exit(1)
	}

	var cli string
	if len(clis) == 1 {
		cli = clis[0]
	} else {
		cli = runSelector(clis)
		if cli == "" {
			return
		}
	}

	if err := launchSession(cli); err != nil {
		fmt.Fprintf(os.Stderr, "Launch failed: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`clipnote — AI CLI output annotation tool (tmux session mode)

Usage:
  clipnote              Launch tmux session (auto-detect AI CLI)
  clipnote --help       Show this help

Keybindings (in annotation panel):
  r       Capture left pane content
  R       Custom range capture
  Ctrl+r  Clear all content and marks
  j/k     Move up/down
  g/G     Jump to top/bottom
  m       Toggle mark
  c       Mark + annotate
  S       Export marks to clipboard
  P       Paste marks to left pane
  [/]     Shrink/expand content panel
  ?       Show help
  q       Quit`)
}
