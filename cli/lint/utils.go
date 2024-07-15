package lint

import (
	"errors"
	"fmt"
	"os"

	term_env "github.com/muesli/termenv"

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v2/pipeline/errors"
)

func FormatLintError(file string, err error) (string, error) {
	if err == nil {
		return "", nil
	}

	output := term_env.NewOutput(os.Stdout)
	str := ""

	amountErrors := 0
	amountWarnings := 0
	linterErrors := pipeline_errors.GetPipelineErrors(err)
	for _, err := range linterErrors {
		line := "  "

		if err.IsWarning {
			line = fmt.Sprintf("%s ⚠️ ", line)
			amountWarnings++
		} else {
			line = fmt.Sprintf("%s ❌", line)
			amountErrors++
		}

		if data := pipeline_errors.GetLinterData(err); data != nil {
			line = fmt.Sprintf("%s %s\t%s", line, output.String(data.Field).Bold(), err.Message)
		} else {
			line = fmt.Sprintf("%s %s", line, err.Message)
		}

		// TODO: use table output
		str = fmt.Sprintf("%s%s\n", str, line)
	}

	if amountErrors > 0 {
		if amountWarnings > 0 {
			str = fmt.Sprintf("🔥 %s has %d errors and warnings:\n%s", output.String(file).Underline(), len(linterErrors), str)
		} else {
			str = fmt.Sprintf("🔥 %s has %d errors:\n%s", output.String(file).Underline(), len(linterErrors), str)
		}
		return str, errors.New("config has errors")
	}

	str = fmt.Sprintf("⚠️  %s has %d warnings:\n%s", output.String(file).Underline(), len(linterErrors), str)
	return str, nil
}
