package main

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up           key.Binding
	Down         key.Binding
	Top          key.Binding
	Bottom       key.Binding
	Mark         key.Binding
	Comment      key.Binding
	Submit       key.Binding
	Quit         key.Binding
	Help         key.Binding
	SubmitNote   key.Binding
	Escape       key.Binding
	ShrinkLeft   key.Binding
	ExpandLeft   key.Binding
	Capture      key.Binding // r — capture visible area (append)
	CaptureRange key.Binding // R — custom range capture (append)
	ClearAll     key.Binding // ctrl+r — clear all content and marks
	PasteToPane  key.Binding // P — paste marks to left pane
}

var keys = KeyMap{
	Up:           key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "up")),
	Down:         key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "down")),
	Top:          key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "top")),
	Bottom:       key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom")),
	Mark:         key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "mark")),
	Comment:      key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "mark+note")),
	Submit:       key.NewBinding(key.WithKeys("S"), key.WithHelp("S", "export")),
	Quit:         key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	Help:         key.NewBinding(key.WithKeys("?", "/"), key.WithHelp("?", "help")),
	SubmitNote:   key.NewBinding(key.WithKeys("ctrl+s")),
	Escape:       key.NewBinding(key.WithKeys("esc")),
	ShrinkLeft:   key.NewBinding(key.WithKeys("["), key.WithHelp("[", "shrink")),
	ExpandLeft:   key.NewBinding(key.WithKeys("]"), key.WithHelp("]", "expand")),
	Capture:      key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "capture visible")),
	CaptureRange: key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "custom range")),
	ClearAll:     key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "clear all")),
	PasteToPane:  key.NewBinding(key.WithKeys("P"), key.WithHelp("P", "paste to left pane")),
}
