package output

import (
	"errors"
	"strings"
)

var ErrOutputOptionRequired = errors.New("output option required")

func ParseOutputOptions(out string) (string, []string) {
	out, opt, found := strings.Cut(out, "=")

	if !found {
		return out, nil
	}

	var optList []string

	if opt != "" {
		optList = strings.Split(opt, ",")
	}

	return out, optList
}
