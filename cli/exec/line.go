package exec

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Identifies the type of line in the logs.
const (
	LineStdout int = iota
	LineStderr
	LineExitCode
	LineMetadata
	LineProgress
)

// Line is a line of console output.
type Line struct {
	Proc string `json:"proc,omitempty"`
	Time int64  `json:"time,omitempty"`
	Type int    `json:"type,omitempty"`
	Pos  int    `json:"pos,omityempty"`
	Out  string `json:"out,omitempty"`
}

// LineWriter sends logs to the client.
type LineWriter struct {
	name  string
	num   int
	now   time.Time
	rep   *strings.Replacer
	lines []*Line
}

// NewLineWriter returns a new line reader.
func NewLineWriter(name string) *LineWriter {
	w := new(LineWriter)
	w.name = name
	w.num = 0
	w.now = time.Now().UTC()

	return w
}

func (w *LineWriter) Write(p []byte) (n int, err error) {
	out := string(p)
	if w.rep != nil {
		out = w.rep.Replace(out)
	}

	line := &Line{
		Out:  out,
		Proc: w.name,
		Pos:  w.num,
		Time: int64(time.Since(w.now).Seconds()),
		Type: LineStdout,
	}

	fmt.Fprintf(os.Stderr, "[%s:L%d:%ds] %s", w.name, w.num, int64(time.Since(w.now).Seconds()), out)

	w.num++

	w.lines = append(w.lines, line)
	return len(p), nil
}
