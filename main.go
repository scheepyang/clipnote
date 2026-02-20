package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
)

func main() {
	// Show usage
	if len(os.Args) >= 2 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		printUsage()
		return
	}

	// IPC client: clipnote ipc <command> [args...]
	if len(os.Args) >= 3 && os.Args[1] == "ipc" {
		command := os.Args[2]
		args := os.Args[3:]
		if err := sendIPCCommand(command, args); err != nil {
			fmt.Fprintf(os.Stderr, "IPC error: %v\n", err)
			os.Exit(1)
		}
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

	// expose program to IPC handler so it can inject messages
	ipcProgram = p

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}
}

func runLauncher() {
	// use CLIPNOTE_CLI env var to skip detection and selector
	if envCLI := os.Getenv("CLIPNOTE_CLI"); envCLI != "" {
		if err := launchSession(envCLI); err != nil {
			fmt.Fprintf(os.Stderr, "Launch failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	clis := detectCLIs()
	if len(clis) == 0 {
		fmt.Fprintln(os.Stderr, "No installed AI CLI found (claude, gemini, codex, aider)")
		os.Exit(1)
	}

	var cli string
	if len(clis) == 1 {
		cli = clis[0]
	} else if isatty.IsTerminal(os.Stdin.Fd()) {
		cli = runSelector(clis)
		if cli == "" {
			return
		}
	} else {
		cli = clis[0]
	}

	if err := launchSession(cli); err != nil {
		fmt.Fprintf(os.Stderr, "Launch failed: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`clipnote -- AI CLI output annotation tool (tmux session mode)

Usage:
  clipnote              Launch tmux session (auto-detect AI CLI)
  clipnote ipc <cmd>    Send IPC command to running annotation TUI
  clipnote --help       Show this help

IPC commands:
  capture               Capture left pane content
  get-marks             Get all marks as JSON
  mark <line> [line...] Mark specific lines (0-indexed)
  export                Export marks to clipboard

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
