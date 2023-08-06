package common

import (
	"strings"
)

func VolumeMountPath(i string) string {
	s := strings.Split(i, ":")
	if len(s) > 1 {
		return s[1]
	}
	return s[0]
}
