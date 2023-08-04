package common

import (
	"os"

	"golang.org/x/term"
)

// IsInteractive checks if the output is piped, but NOT if the session is run interactively..
func IsInteractive() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
