---
name: launch
description: Launch clipnote tmux annotation session. Use when the user
  explicitly mentions clipnote, wants to annotate/mark/highlight AI output,
  wants to review AI responses or take notes, wants a split-pane annotation
  panel, or wants to export/summarize key points from AI output.
---

# clipnote:launch

Launch the clipnote annotation session via tmux.

## Instructions

1. When you detect the user's intent matches this skill, first explain:
   - If running inside tmux: clipnote will open as a **split pane** in the current window (no context switch)
   - If not inside tmux: clipnote will open in a **new Terminal.app window**
   - Multiple launches will reuse the existing pane instead of creating new ones
2. Ask the user to confirm before proceeding (e.g. "Shall I launch clipnote now?")
3. Only after the user confirms, run the following command using the Bash tool:

```bash
CLIPNOTE_CLI=claude "${CLAUDE_PLUGIN_ROOT}/bin/clipnote"
```

When inside tmux, this splits the current window with the annotation panel on the right (45% width).
When outside tmux, this opens a new Terminal.app window with a full tmux session.

Note: `CLIPNOTE_CLI=claude` bypasses the interactive CLI selector, which cannot run inside Claude Code's non-TTY environment.

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
