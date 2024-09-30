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
	"io"
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

	peer      rpc.Peer
	stepUUID  string
	num       int
	startTime time.Time
	replacer  *strings.Replacer
}

// NewLineWriter returns a new line reader.
func NewLineWriter(peer rpc.Peer, stepUUID string, secret ...string) io.Writer {
	lw := &LineWriter{
		peer:      peer,
		stepUUID:  stepUUID,
		startTime: time.Now().UTC(),
		replacer:  shared.NewSecretsReplacer(secret),
	}
	return lw
}

func (w *LineWriter) Write(p []byte) (n int, err error) {
	data := string(p)
	if w.replacer != nil {
		data = w.replacer.Replace(data)
	}
	log.Trace().Str("step-uuid", w.stepUUID).Msgf("grpc write line: %s", data)

	line := &rpc.LogEntry{
		Data:     []byte(strings.TrimSuffix(data, "\n")), // remove trailing newline
		StepUUID: w.stepUUID,
		Time:     int64(time.Since(w.startTime).Seconds()),
		Type:     rpc.LogEntryStdout,
		Line:     w.num,
	}

	w.num++

	w.peer.EnqueueLog(line)
	return len(data), nil
}
