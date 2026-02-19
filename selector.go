package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type selectorModel struct {
	choices  []string
	cursor   int
	selected string
}

func (m selectorModel) Init() tea.Cmd {
	return nil
}

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.choices[m.cursor]
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m selectorModel) View() string {
	var b strings.Builder
	b.WriteString("Multiple AI CLIs detected. Select one:\n\n")

	for i, choice := range m.choices {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, choice))
	}

	b.WriteString("\nj/k or arrows to move, Enter to select, q to cancel")
	return b.String()
}

func runSelector(choices []string) string {
	m := selectorModel{choices: choices}
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Selector error: %v\n", err)
		return ""
	}
	return result.(selectorModel).selected
}
