package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const ipcSocketPath = "/tmp/clipnote.sock"

// IPC message types (controller -> TUI)
type ipcRequest struct {
	Type  string `json:"type"`
	Lines []int  `json:"lines,omitempty"`
}

// IPC response types (TUI -> controller)
type ipcResponse struct {
	Type    string      `json:"type"`
	Data    any `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// IPCMsg delivers an IPC request into the bubbletea event loop
type IPCMsg struct {
	Request  ipcRequest
	ReplyCh  chan<- ipcResponse
}

// startIPCServer creates a Unix socket server and returns a tea.Cmd
// that listens for incoming connections. Each request is forwarded
// to the bubbletea event loop via IPCMsg.
func startIPCServer() tea.Cmd {
	return func() tea.Msg {
		// clean up stale socket
		os.Remove(ipcSocketPath)

		ln, err := net.Listen("unix", ipcSocketPath)
		if err != nil {
			return nil
		}

		// accept connections in background
		go func() {
			defer ln.Close()
			defer os.Remove(ipcSocketPath)

			for {
				conn, err := ln.Accept()
				if err != nil {
					return
				}
				go handleIPCConn(conn)
			}
		}()

		return nil
	}
}

// ipcProgram is set by the TUI so the IPC handler can inject messages
var ipcProgram *tea.Program

func handleIPCConn(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var req ipcRequest
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			writeJSON(conn, ipcResponse{Type: "error", Message: "invalid JSON"})
			continue
		}

		replyCh := make(chan ipcResponse, 1)

		if ipcProgram != nil {
			ipcProgram.Send(IPCMsg{Request: req, ReplyCh: replyCh})
		} else {
			replyCh <- ipcResponse{Type: "error", Message: "TUI not ready"}
		}

		// wait for reply from the bubbletea loop
		resp := <-replyCh
		writeJSON(conn, resp)
	}
}

func writeJSON(conn net.Conn, v any) {
	data, _ := json.Marshal(v)
	data = append(data, '\n')
	conn.Write(data)
}

// handleIPC processes an IPC request within the model and returns a response.
func (m *Model) handleIPC(req ipcRequest) ipcResponse {
	switch req.Type {
	case "capture":
		return m.ipcCapture()
	case "mark":
		return m.ipcMark(req.Lines)
	case "get-marks":
		return m.ipcGetMarks()
	case "export":
		return m.ipcExport()
	default:
		return ipcResponse{Type: "error", Message: fmt.Sprintf("unknown command: %s", req.Type)}
	}
}

func (m *Model) ipcCapture() ipcResponse {
	out, err := execCommand("tmux", "capture-pane", "-p", "-t", m.tmuxPane).CombinedOutput()
	if err != nil {
		return ipcResponse{Type: "error", Message: "capture failed: " + strings.TrimSpace(string(out))}
	}

	content := strings.TrimRight(string(out), "\n")
	newLines := strings.Split(content, "\n")

	m.captureCount++
	if len(m.lines) > 0 {
		separator := fmt.Sprintf("─── Capture #%d ───", m.captureCount)
		m.lines = append(m.lines, separator)
	}
	jumpTo := len(m.lines)
	m.lines = append(m.lines, newLines...)
	m.cursorLine = jumpTo
	m.syncViewport()
	m.statusMsg = fmt.Sprintf("Captured %s lines (total %s)", itoa(len(newLines)), itoa(len(m.lines)))

	return ipcResponse{
		Type: "result",
		Data: map[string]any{
			"lines_captured": len(newLines),
			"total_lines":    len(m.lines),
		},
	}
}

func (m *Model) ipcMark(lines []int) ipcResponse {
	if len(lines) == 0 {
		return ipcResponse{Type: "error", Message: "no lines specified"}
	}

	marked := 0
	for _, line := range lines {
		if line < 0 || line >= len(m.lines) {
			continue
		}
		if !m.HasMark(line) {
			m.ToggleMark(line)
			marked++
		}
	}
	m.statusMsg = fmt.Sprintf("Marked %d lines via IPC", marked)

	return ipcResponse{
		Type: "result",
		Data: map[string]any{
			"marked": marked,
		},
	}
}

type markData struct {
	Line int    `json:"line"`
	Text string `json:"text"`
	Note string `json:"note,omitempty"`
}

func (m *Model) ipcGetMarks() ipcResponse {
	marks := make([]markData, len(m.marks))
	for i, mk := range m.marks {
		marks[i] = markData{Line: mk.Line, Text: mk.Text, Note: mk.Note}
	}
	return ipcResponse{Type: "result", Data: marks}
}

func (m *Model) ipcExport() ipcResponse {
	text := m.ExportMarks()
	if text == "" {
		return ipcResponse{Type: "result", Data: map[string]string{"exported": ""}}
	}

	msg := m.CopyMarksToClipboard()
	return ipcResponse{
		Type: "result",
		Data: map[string]any{
			"exported": text,
			"status":   msg,
		},
	}
}

// sendIPCCommand connects to the IPC socket, sends a command, and prints the response.
// Used by the CLI client (clipnote ipc <command>).
func sendIPCCommand(command string, args []string) error {
	conn, err := net.Dial("unix", ipcSocketPath)
	if err != nil {
		return fmt.Errorf("cannot connect to clipnote (is annotation TUI running?): %w", err)
	}
	defer conn.Close()

	req := ipcRequest{Type: command}

	// parse additional args for mark command
	if command == "mark" && len(args) > 0 {
		var lines []int
		for _, arg := range args {
			var n int
			if _, err := fmt.Sscanf(arg, "%d", &n); err == nil {
				lines = append(lines, n)
			}
		}
		req.Lines = lines
	}

	data, _ := json.Marshal(req)
	data = append(data, '\n')
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	return nil
}
