package rpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Identifies the type of line in the logs.
type LineType int

const (
	LineStdout LineType = iota
	LineStderr
	LineExitCode
	LineMetadata
	LineProgress
)

// Line is a line of console output.
type Line struct {
	Proc string   `json:"proc,omitempty"`
	Time int64    `json:"time,omitempty"`
	Type LineType `json:"type,omitempty"`
	Pos  int      `json:"pos,omitempty"`
	Out  string   `json:"out,omitempty"`
}

func (l *Line) String() string {
	switch l.Type {
	case LineExitCode:
		return fmt.Sprintf("[%s] exit code %s", l.Proc, l.Out)
	default:
		return fmt.Sprintf("[%s:L%v:%vs] %s", l.Proc, l.Pos, l.Time, l.Out)
	}
}

// LineWriter sends logs to the client.
type LineWriter struct {
	peer     Peer
	id       string
	name     string
	num      int
	now      time.Time
	replacer *strings.Replacer
}

// NewLineWriter returns a new line reader.
func NewLineWriter(peer Peer, id, name string, secrets ...string) *LineWriter {
	w := new(LineWriter)
	w.peer = peer
	w.id = id
	w.name = name
	w.num = 0
	w.now = time.Now().UTC()

	var oldnewSecrets []string
	for _, secret := range secrets {
		oldnewSecrets = append(oldnewSecrets, secret)
		oldnewSecrets = append(oldnewSecrets, "********")
	}
	if len(oldnewSecrets) != 0 {
		w.replacer = strings.NewReplacer(oldnewSecrets...)
	}

	return w
}

func (w *LineWriter) Write(p []byte) (n int, err error) {
	data := string(p)
	if w.replacer != nil {
		data = w.replacer.Replace(data)
	}

	line := &Line{
		Out:  data,
		Proc: w.name,
		Pos:  w.num,
		Time: int64(time.Since(w.now).Seconds()),
		Type: LineStdout,
	}
	if err := w.peer.Log(context.Background(), w.id, line); err != nil {
		log.Error().Err(err).Msgf("fail to write pipeline log to peer '%s'", w.id)
	}

	if strings.Contains(data, "\n") {
		w.num++
	}

	return len(p), nil
}
