package ui

import (
	"charm.land/huh/v2"
)

func Confirm(prompt string) (bool, error) {
	var confirm bool
	err := huh.NewConfirm().
		Title(prompt).
		Affirmative("Yes!").
		Negative("No.").
		Value(&confirm).Run()
	if err != nil {
		return false, err
	}

	return confirm, err
}
