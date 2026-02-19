package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

var execCommand = exec.Command

const noteInputHeight = 4

// CaptureAppendMsg is the message type for capture results
type CaptureAppendMsg string

type overlayKind int

const (
	overlayNone overlayKind = iota
	overlayHelp
	overlayNote
)

type Model struct {
	lines     []string
	noteInput textarea.Model
	marks      []Mark
	cursorLine int
	inputMode  bool
	overlayType overlayKind
	width      int
	height     int
	statusMsg  string
	ready      bool
	splitRatio   int // left content panel width percentage (default 70)
	scrollOffset int // manually managed scroll offset

	tmuxPane       string // left pane tmux ID (e.g. clipnote:0.0)
	captureCount   int    // capture counter for separator lines (#N)
	captureInput   bool   // whether R line count input mode is active
	captureInputBuf string // R input buffer text
	captureConfirm bool   // whether confirming full scrollback capture
}

func NewWatchModel(paneID string) Model {
	ta := textarea.New()
	ta.Placeholder = "Enter note..."
	ta.CharLimit = 500
	ta.SetHeight(noteInputHeight)
	ta.ShowLineNumbers = false

	return Model{
		lines:      []string{},
		noteInput:  ta,
		marks:      []Mark{},
		splitRatio: 70,
		tmuxPane:   paneID,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

// captureVisible captures the visible area of the left pane (scroll-position aware)
func captureVisible(paneID string) tea.Cmd {
	return func() tea.Msg {
		scrollPos := tmuxDisplayVar(paneID, "scroll_position")
		if scrollPos == "" || scrollPos == "0" {
			return captureExec(paneID, nil)
		}

		sp, err := strconv.Atoi(scrollPos)
		if err != nil {
			return captureExec(paneID, nil)
		}

		ph, err := strconv.Atoi(tmuxDisplayVar(paneID, "pane_height"))
		if err != nil || ph <= 0 {
			return captureExec(paneID, nil)
		}

		// visible top = -sp, visible bottom = ph - 1 - sp
		startLine := -sp
		endLine := ph - 1 - sp
		return captureExec(paneID, []string{
			"-S", strconv.Itoa(startLine),
			"-E", strconv.Itoa(endLine),
		})
	}
}

// captureRange captures N lines from the left pane (scroll-position aware)
func captureRange(paneID string, lines int) tea.Cmd {
	return func() tea.Msg {
		scrollPos := tmuxDisplayVar(paneID, "scroll_position")
		if scrollPos == "" || scrollPos == "0" {
			// not scrolled, capture n lines from bottom
			return captureExec(paneID, []string{"-S", fmt.Sprintf("-%d", lines)})
		}

		sp, err := strconv.Atoi(scrollPos)
		if err != nil {
			return captureExec(paneID, []string{"-S", fmt.Sprintf("-%d", lines)})
		}

		ph, err2 := strconv.Atoi(tmuxDisplayVar(paneID, "pane_height"))
		if err2 != nil || ph <= 0 {
			return captureExec(paneID, []string{"-S", fmt.Sprintf("-%d", lines)})
		}

		// visible bottom = ph - 1 - sp, capture N lines upward
		endLine := ph - 1 - sp
		startLine := endLine - lines + 1
		return captureExec(paneID, []string{
			"-S", strconv.Itoa(startLine),
			"-E", strconv.Itoa(endLine),
		})
	}
}

// captureAll captures the entire scrollback of the left pane
func captureAll(paneID string) tea.Cmd {
	return capturePane(paneID, "-")
}

// capturePane runs tmux capture-pane and returns the result
func capturePane(paneID, startLine string) tea.Cmd {
	return func() tea.Msg {
		args := []string{"capture-pane", "-p", "-t", paneID}
		if startLine != "" {
			args = append(args, "-S", startLine)
		}

		cmd := execCommand("tmux", args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			errMsg := strings.TrimSpace(string(out))
			if errMsg == "" {
				errMsg = err.Error()
			}
			return CaptureAppendMsg("Capture failed:\n" + errMsg)
		}

		// trim trailing empty lines
		result := strings.TrimRight(string(out), "\n")
		return CaptureAppendMsg(result)
	}
}

// tmuxDisplayVar queries a tmux pane format variable
func tmuxDisplayVar(paneID, varName string) string {
	out, err := execCommand("tmux", "display-message",
		"-t", paneID, "-p", "#{"+varName+"}").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// captureExec runs tmux capture-pane and returns the result (used by captureVisible)
func captureExec(paneID string, extraArgs []string) tea.Msg {
	args := []string{"capture-pane", "-p", "-t", paneID}
	args = append(args, extraArgs...)
	out, err := execCommand("tmux", args...).CombinedOutput()
	if err != nil {
		errMsg := strings.TrimSpace(string(out))
		if errMsg == "" {
			errMsg = err.Error()
		}
		return CaptureAppendMsg("Capture failed:\n" + errMsg)
	}
	return CaptureAppendMsg(strings.TrimRight(string(out), "\n"))
}
