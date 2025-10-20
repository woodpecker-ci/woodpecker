package ui

import (
	"errors"
	"strings"

	"github.com/charmbracelet/huh"
)

func Ask(prompt, placeholder string, required bool) (string, error) {
	var input string
	err := huh.NewInput().
		Title(prompt).
		Value(&input).
		Placeholder(placeholder).Validate(func(s string) error {
		if required && strings.TrimSpace(s) == "" {
			return errors.New("required")
		}
		return nil
	}).Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}
