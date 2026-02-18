package main

import (
	"fmt"
	"os"
	"os/exec"
)

var knownCLIs = []string{"claude", "gemini", "codex", "aider"}

func isInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

func detectCLIs() []string {
	var found []string
	for _, cli := range knownCLIs {
		if _, err := exec.LookPath(cli); err == nil {
			found = append(found, cli)
		}
	}
	return found
}

func launchSession(cli string) error {
	sessionName := "ai-review"
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// kill any leftover session with the same name
	exec.Command("tmux", "kill-session", "-t", sessionName).Run()

	// create session, left pane runs the CLI
	cmd := exec.Command("tmux",
		"new-session", "-d", "-s", sessionName, "-x", "200", "-y", "50", cli)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create tmux session: %w\n%s", err, out)
	}

	// right pane launches annotation TUI (45% width)
	splitCmd := exec.Command("tmux",
		"split-window", "-h", "-t", sessionName, "-l", "45%",
		self, "--internal-watch", sessionName+":0.0")
	if out, err := splitCmd.CombinedOutput(); err != nil {
		exec.Command("tmux", "kill-session", "-t", sessionName).Run()
		return fmt.Errorf("failed to split pane: %w\n%s", err, out)
	}

	// enable mouse support + unified pane border color
	exec.Command("tmux", "set-option", "-t", sessionName, "mouse", "on").Run()
	exec.Command("tmux", "set-option", "-t", sessionName, "pane-border-style", "fg=colour62").Run()
	exec.Command("tmux", "set-option", "-t", sessionName, "pane-active-border-style", "fg=colour62").Run()

	// attach to session (blocks until user detaches)
	attachCmd := exec.Command("tmux", "attach-session", "-t", sessionName)
	attachCmd.Stdin = os.Stdin
	attachCmd.Stdout = os.Stdout
	attachCmd.Stderr = os.Stderr
	return attachCmd.Run()
}
