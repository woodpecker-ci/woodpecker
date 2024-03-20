package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type confirmModel struct {
	confirmed bool
	prompt    string
	err       error
}

func (m confirmModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Runes != nil {
			switch msg.Runes[0] {
			case 'y':
				m.confirmed = true
				return m, tea.Quit
			case 'n':
				m.confirmed = false
				return m, tea.Quit
			}
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	default:
		return m, nil
	}

	return m, cmd
}

func (m confirmModel) View() string {
	return fmt.Sprintf(
		"%s y / n (esc to quit)",
		m.prompt,
	) + "\n"
}

func Confirm(prompt string) (bool, error) {
	p := tea.NewProgram(confirmModel{
		prompt: prompt,
		err:    nil,
	})

	_m, err := p.Run()
	if err != nil {
		return false, err
	}

	m, ok := _m.(confirmModel)
	if !ok {
		return false, fmt.Errorf("unexpected model: %T", _m)
	}

	return m.confirmed, nil
}
