---
name: annotate
description: Programmatically operate the clipnote annotation TUI via IPC commands.
  Use when you need to capture AI output, get existing marks, mark specific lines,
  or export marks -- all without requiring user interaction with the TUI.
  Requires the annotation TUI to be running (launched via the "launch" skill).
---

# clipnote:annotate

Operate the clipnote annotation TUI programmatically via IPC.

## Prerequisites

The annotation TUI must already be running. If it is not, use the `launch` skill first.

## Available Commands

All commands use the `clipnote ipc` subcommand and return NDJSON responses.

### Capture left pane content

Captures the current visible content from the left tmux pane into the annotation TUI.

```bash
"${CLAUDE_PLUGIN_ROOT}/bin/clipnote" ipc capture
```

Response:
```json
{"type":"result","data":{"lines_captured":42,"total_lines":42}}
```

### Get all marks

Returns all current marks with their line numbers, text, and notes.

```bash
"${CLAUDE_PLUGIN_ROOT}/bin/clipnote" ipc get-marks
```

Response:
```json
{"type":"result","data":[{"line":5,"text":"some code here","note":"important"}]}
```

### Mark specific lines

Mark one or more lines by 0-indexed line number. Lines that are already marked will be skipped.

```bash
"${CLAUDE_PLUGIN_ROOT}/bin/clipnote" ipc mark 5 6 7
```

Response:
```json
{"type":"result","data":{"marked":3}}
```

### Export marks to clipboard

Exports all marks to the system clipboard and returns the exported text.

```bash
"${CLAUDE_PLUGIN_ROOT}/bin/clipnote" ipc export
```

Response:
```json
{"type":"result","data":{"exported":"marked text here\n> [Q] note","status":"Copied 2 marks to clipboard"}}
```

## Error Handling

If the TUI is not running, commands will fail with a connection error.
If a command is invalid, the response will be:

```json
{"type":"error","message":"unknown command: foo"}
```

## Typical Workflow

1. Ensure TUI is running (use `launch` skill if needed)
2. Run `clipnote ipc capture` to grab the current AI output
3. Run `clipnote ipc get-marks` to check what the user has marked
4. Optionally run `clipnote ipc export` to get the marked content
