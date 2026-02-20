package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
)

var knownCLIs = []string{"claude", "gemini", "codex", "aider"}

const paneIDFile = "/tmp/clipnote-pane-id"

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
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		return launchDetached(cli)
	}
	return launchInCurrentTerminal(cli)
}

// launchDetached launches clipnote without a TTY.
// If already inside tmux, split-window in the current session (same window).
// Otherwise, open a new Terminal.app window via osascript (macOS fallback).
func launchDetached(cli string) error {
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	if isInsideTmux() {
		return launchInTmuxSplit(self, cli)
	}

	// fallback: open a new Terminal.app window via osascript (macOS only)
	script := fmt.Sprintf("CLIPNOTE_CLI=%s %s", cli, self)
	cmd := exec.Command("osascript", "-e",
		fmt.Sprintf(`tell application "Terminal"
activate
do script "%s"
end tell`, script))
	return cmd.Run()
}

// launchInTmuxSplit opens clipnote in a split pane within the current tmux window.
// Reuses an existing pane if one is still alive.
func launchInTmuxSplit(self, cli string) error {
	launchCmd := fmt.Sprintf("CLIPNOTE_CLI=%s %s", cli, self)

	// try to reuse existing pane
	if reused := tryReusePane(launchCmd); reused {
		return nil
	}

	// create new split pane (45% width)
	splitCmd := exec.Command("tmux", "split-window", "-h", "-l", "45%",
		"-P", "-F", "#{pane_id}", launchCmd)
	out, err := splitCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux split-window failed: %w\n%s", err, out)
	}

	savePaneID(strings.TrimSpace(string(out)))
	return nil
}

// launchInCurrentTerminal creates a tmux session directly in the current terminal
func launchInCurrentTerminal(cli string) error {
	sessionName := "clipnote"
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
		"-P", "-F", "#{pane_id}",
		self, "--internal-watch", sessionName+":0.0")
	out, err2 := splitCmd.CombinedOutput()
	if err2 != nil {
		exec.Command("tmux", "kill-session", "-t", sessionName).Run()
		return fmt.Errorf("failed to split pane: %w\n%s", err2, out)
	}

	savePaneID(strings.TrimSpace(string(out)))

	// enable mouse support + unified pane border color
	exec.Command("tmux", "set-option", "-t", sessionName, "mouse", "on").Run()
	exec.Command("tmux", "set-option", "-t", sessionName, "pane-border-style", "fg=colour62").Run()
	exec.Command("tmux", "set-option", "-t", sessionName, "pane-active-border-style", "fg=colour62").Run()

	// bind prefix+a to reopen annotation pane if closed
	reopenCmd := fmt.Sprintf("%s --internal-watch %s:0.0", self, sessionName)
	exec.Command("tmux", "bind-key", "a",
		"split-window", "-h", "-t", sessionName+":0.0", "-l", "45%",
		"sh", "-c", reopenCmd).Run()

	// attach to session (blocks until user detaches)
	attachCmd := exec.Command("tmux", "attach-session", "-t", sessionName)
	attachCmd.Stdin = os.Stdin
	attachCmd.Stdout = os.Stdout
	attachCmd.Stderr = os.Stderr
	return attachCmd.Run()
}

// tryReusePane checks if a saved pane is still alive and sends a new command to it.
// Returns true if the pane was successfully reused.
func tryReusePane(launchCmd string) bool {
	paneID := loadPaneID()
	if paneID == "" {
		return false
	}

	if !isPaneAlive(paneID) {
		os.Remove(paneIDFile)
		return false
	}

	// send Ctrl+C to interrupt any running process
	exec.Command("tmux", "send-keys", "-t", paneID, "C-c", "").Run()
	time.Sleep(150 * time.Millisecond)

	// send the new launch command
	exec.Command("tmux", "send-keys", "-t", paneID, launchCmd, "Enter").Run()
	return true
}

// isPaneAlive checks if a tmux pane still exists
func isPaneAlive(paneID string) bool {
	out, err := exec.Command("tmux", "display-message",
		"-t", paneID, "-p", "#{pane_id}").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

func savePaneID(id string) {
	os.WriteFile(paneIDFile, []byte(id), 0644)
}

func loadPaneID() string {
	data, err := os.ReadFile(paneIDFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
