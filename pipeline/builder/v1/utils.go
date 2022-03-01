package v1

import (
	"path/filepath"
	"strings"
)

func sanitizePipelinePath(path string) string {
	path = filepath.Base(path)
	path = strings.TrimSuffix(path, ".yml")
	path = strings.TrimPrefix(path, ".")
	return path
}
