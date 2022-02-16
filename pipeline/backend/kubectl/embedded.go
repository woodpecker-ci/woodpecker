package kubectl

import (
	"embed"
)

//go:embed templates/*.yaml

var Embedded embed.FS
