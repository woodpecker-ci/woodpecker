// Copyright 2022 Woodpecker Authors
// Copyright 2011 Drone.IO Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/shared"
)

// LineWriter sends logs to the client.
type LineWriter struct {
	sync.Mutex

	peer          rpc.Peer
	stepUUID      string
	num           int
	startTime     time.Time
	replacer      *strings.Replacer
	pendingLines  []*rpc.LogEntry
	size          int
	bufferSize    int
	flushInterval time.Duration
	timer         *time.Timer
	closeChan     chan struct{}
}

// NewLineWriter returns a new line reader.
func NewLineWriter(peer rpc.Peer, stepUUID string, flushInterval time.Duration, secret ...string) *LineWriter {
	lw := &LineWriter{
		peer:          peer,
		stepUUID:      stepUUID,
		startTime:     time.Now().UTC(),
		replacer:      shared.NewSecretsReplacer(secret),
		pendingLines:  nil,
		flushInterval: flushInterval,
		timer:         time.NewTimer(flushInterval),
		closeChan:     make(chan struct{}),
	}
	go lw.start()
	return lw
}

func (w *LineWriter) Write(p []byte) (n int, err error) {
	s := string(p)
	v := []string{s}

	// if the string contains a newline character, we split
	if strings.Contains(strings.TrimSuffix(s, "\n"), "\n") {
		v = strings.SplitAfter(s, "\n")
	}

	for _, line := range v {
		n, err := w.writeLine(line)
		if err != nil {
			return n, err
		}
	}

	return len(p), nil
}

func (w *LineWriter) writeLine(data string) (n int, err error) {
	if w.replacer != nil {
		data = w.replacer.Replace(data)
	}
	log.Trace().Str("step-uuid", w.stepUUID).Msgf("grpc write line: %s", data)

	line := &rpc.LogEntry{
		Data:     strings.TrimSpace(data),
		StepUUID: w.stepUUID,
		Time:     int64(time.Since(w.startTime).Seconds()),
		Type:     rpc.LogEntryStdout,
		Line:     w.num,
	}

	w.num++
	w.size += len(data)

	if w.size > w.bufferSize {
		if err := w.flush(); err != nil {
			return 0, err
		}
	}

	w.pendingLines = append(w.pendingLines, line)
	return len(data), nil
}

func (w *LineWriter) start() {
	for {
		select {
		case <-w.closeChan:
			return
		case <-time.After(w.flushInterval):
			if err := w.flush(); err != nil {
				log.Error().Str("step-uuid", w.stepUUID).Err(err).Msg("flushing log entries")
			}
		}
	}
}

func (w *LineWriter) flush() error {
	pendingLines := []*rpc.LogEntry{}

	w.Lock()
	pendingLines = append(pendingLines, w.pendingLines...)
	w.pendingLines = nil
	defer w.Unlock()

	for _, line := range pendingLines {
		// TODO: send log entries in batch
		if err := w.peer.Log(context.Background(), line); err != nil {
			return err
		}
	}

	return nil
}

func (w *LineWriter) Close() error {
	w.timer.Stop()
	close(w.closeChan)
	return w.flush()
}
