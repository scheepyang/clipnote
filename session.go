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

func launchSession(cli, sessionID string) error {
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		return launchDetached(cli, sessionID)
	}
	return launchInCurrentTerminal(cli, sessionID)
}

// launchDetached launches clipnote without a TTY.
// If already inside tmux, split-window in the current session (same window).
// Otherwise, create a new tmux session with resume support and attach via Terminal.app.
func launchDetached(cli, sessionID string) error {
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	if isInsideTmux() {
		return launchInTmuxSplit(self)
	}

	// create detached tmux session, left pane runs CLI (with resume support)
	return launchNewTmuxAndAttach(cli, sessionID, self)
}

// currentPaneID returns the tmux pane ID of the caller (e.g. Claude Code's pane)
func currentPaneID() string {
	out, err := exec.Command("tmux", "display-message", "-p", "#{pane_id}").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// launchInTmuxSplit opens the annotation TUI in a split pane within the current tmux window.
// The TUI watches the caller's pane (Claude Code) directly — no new tmux session is created.
// Reuses an existing pane if one is still alive.
func launchInTmuxSplit(self string) error {
	// get the pane ID of the caller (Claude Code) so the TUI can watch it
	callerPane := currentPaneID()
	if callerPane == "" {
		return fmt.Errorf("failed to detect current tmux pane")
	}

	// the split pane runs the annotation TUI directly, watching the caller's pane
	watchCmd := fmt.Sprintf("%s --internal-watch %s", self, callerPane)

	// try to reuse existing pane
	if reused := tryReusePane(watchCmd); reused {
		return nil
	}

	// create new split pane (45% width), run annotation TUI directly
	splitCmd := exec.Command("tmux", "split-window", "-h", "-l", "45%",
		"-P", "-F", "#{pane_id}", watchCmd)
	out, err := splitCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux split-window failed: %w\n%s", err, out)
	}

	savePaneID(strings.TrimSpace(string(out)))

	// bind prefix+a to reopen annotation pane if closed
	reopenCmd := fmt.Sprintf("%s --internal-watch %s", self, callerPane)
	exec.Command("tmux", "bind-key", "a",
		"split-window", "-h", "-t", callerPane, "-l", "45%",
		"sh", "-c", reopenCmd).Run()

	return nil
}

// cliCommand returns the shell command for the left pane.
// If sessionID is provided, uses `claude --resume <id>` to restore the conversation.
func cliCommand(cli, sessionID string) string {
	if sessionID != "" && cli == "claude" {
		return fmt.Sprintf("claude --resume %s", sessionID)
	}
	return cli
}

// launchInCurrentTerminal creates a tmux session directly in the current terminal
func launchInCurrentTerminal(cli, sessionID string) error {
	sessionName := "clipnote"
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// kill any leftover session with the same name
	exec.Command("tmux", "kill-session", "-t", sessionName).Run()

	// create session, left pane runs the CLI (with resume if session ID provided)
	leftCmd := cliCommand(cli, sessionID)
	cmd := exec.Command("tmux",
		"new-session", "-d", "-s", sessionName, "-x", "200", "-y", "50", leftCmd)
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

	configureTmuxSession(sessionName, self)

	// attach to session (blocks until user detaches)
	attachCmd := exec.Command("tmux", "attach-session", "-t", sessionName)
	attachCmd.Stdin = os.Stdin
	attachCmd.Stdout = os.Stdout
	attachCmd.Stderr = os.Stderr
	return attachCmd.Run()
}

// launchNewTmuxAndAttach creates a detached tmux session and opens the user's terminal to attach.
// Used when not inside tmux and no TTY (plugin subprocess).
func launchNewTmuxAndAttach(cli, sessionID, self string) error {
	sessionName := "clipnote"

	// kill any leftover session with the same name
	exec.Command("tmux", "kill-session", "-t", sessionName).Run()

	// create detached session, left pane runs CLI with resume
	leftCmd := cliCommand(cli, sessionID)
	cmd := exec.Command("tmux",
		"new-session", "-d", "-s", sessionName, "-x", "200", "-y", "50", leftCmd)
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

	configureTmuxSession(sessionName, self)

	// open a terminal window via osascript to attach to the tmux session
	osascriptCmd := exec.Command("osascript", "-e", terminalAttachScript(sessionName))
	return osascriptCmd.Run()
}

// terminalAttachScript returns an AppleScript that opens the user's terminal
// and attaches to the given tmux session. Detects the terminal via TERM_PROGRAM.
func terminalAttachScript(sessionName string) string {
	attachCmd := fmt.Sprintf("tmux attach-session -t %s", sessionName)
	switch os.Getenv("TERM_PROGRAM") {
	case "iTerm.app":
		return fmt.Sprintf(`tell application "iTerm"
activate
create window with default profile command "%s"
end tell`, attachCmd)
	default:
		return fmt.Sprintf(`tell application "Terminal"
activate
do script "%s"
end tell`, attachCmd)
	}
}

// configureTmuxSession sets mouse, border color, and bind-key for the session.
func configureTmuxSession(sessionName, self string) {
	// enable mouse support + unified pane border color
	exec.Command("tmux", "set-option", "-t", sessionName, "mouse", "on").Run()
	exec.Command("tmux", "set-option", "-t", sessionName, "pane-border-style", "fg=colour62").Run()
	exec.Command("tmux", "set-option", "-t", sessionName, "pane-active-border-style", "fg=colour62").Run()

	// bind prefix+a to reopen annotation pane if closed
	reopenCmd := fmt.Sprintf("%s --internal-watch %s:0.0", self, sessionName)
	exec.Command("tmux", "bind-key", "a",
		"split-window", "-h", "-t", sessionName+":0.0", "-l", "45%",
		"sh", "-c", reopenCmd).Run()
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
