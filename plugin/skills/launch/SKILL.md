---
name: launch
description: Launch clipnote tmux annotation session. Use when
  the user explicitly asks to start, open, or launch clipnote.
---

# clipnote:launch

Launch the clipnote annotation session via tmux.

## Instructions

Run the following command using the Bash tool:

```bash
"${CLAUDE_PLUGIN_ROOT}/bin/clipnote"
```

This will:
1. Auto-detect installed AI CLIs (claude, gemini, codex, aider)
2. If multiple CLIs are found, show a selection menu
3. Create a tmux session with the AI CLI on the left and the annotation panel on the right

## Keybindings (tell the user)

| Key | Action |
|-----|--------|
| r | Capture visible area |
| R | Custom range capture |
| Ctrl+r | Clear all content |
| j/k | Move cursor up/down |
| g/G | Jump to top/bottom |
| m | Toggle mark |
| c | Mark + add note |
| v | View note |
| S | Export marks to clipboard |
| P | Paste marks to left pane |
| [/] | Shrink/expand content panel |
| ? | Show help |
| q | Quit |
