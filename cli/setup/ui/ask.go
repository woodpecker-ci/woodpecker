package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type askModel struct {
	prompt    string
	required  bool
	textInput textinput.Model
	err       error
}

func (m askModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m askModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if !m.required || (m.required && strings.TrimSpace(m.textInput.Value()) != "") {
				return m, tea.Quit
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	default:
		return m, cmd
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m askModel) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		m.prompt,
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}

func Ask(prompt, placeholder string, required bool) (string, error) {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40

	p := tea.NewProgram(askModel{
		prompt:    prompt,
		textInput: ti,
		required:  required,
		err:       nil,
	})

	_m, err := p.Run()
	if err != nil {
		return "", err
	}

	m, ok := _m.(askModel)
	if !ok {
		return "", fmt.Errorf("unexpected model: %T", _m)
	}

	text := strings.TrimSpace(m.textInput.Value())

	return text, nil
}
