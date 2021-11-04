package template

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

// format formats a template using gofmt.
func format(in io.Reader) (io.Reader, error) {
	var out bytes.Buffer

	gofmt := exec.Command("gofmt", "-s")
	gofmt.Stdin = in
	gofmt.Stdout = &out
	gofmt.Stderr = os.Stderr
	err := gofmt.Run()
	return &out, err
}
