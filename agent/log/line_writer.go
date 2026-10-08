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
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/rpc"
)

// LineWriter sends logs to the client.
type LineWriter struct {
	sync.Mutex

	peer      rpc.Peer
	stepUUID  string
	num       int
	startTime time.Time
}

// NewLineWriter returns a new line reader.
//
// Secret masking is not handled here; wrap the writer with
// shared.NewSecretsReplaceWriter.
func NewLineWriter(peer rpc.Peer, stepUUID string) io.Writer {
	lw := &LineWriter{
		peer:      peer,
		stepUUID:  stepUUID,
		startTime: time.Now().UTC(),
	}
	return lw
}

func (w *LineWriter) Write(p []byte) (n int, err error) {
	log.Trace().Str("step-uuid", w.stepUUID).Msgf("grpc write line: %s", p)

	line := &rpc.LogEntry{
		// remove trailing newline; clone as p must not be retained
		Data:     bytes.Clone(bytes.TrimSuffix(p, []byte("\n"))),
		StepUUID: w.stepUUID,
		Time:     int64(time.Since(w.startTime).Seconds()),
		Type:     rpc.LogEntryStdout,
		Line:     w.num,
	}

	w.num++

	w.peer.EnqueueLog(line)
	return len(p), nil
}
