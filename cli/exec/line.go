// Copyright 2022 Woodpecker Authors
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

package exec

import (
	"fmt"
	"io"
	"os"
	"time"
)

// LineWriter sends logs to the client.
type LineWriter struct {
	stepName  string
	stepUUID  string
	num       int
	startTime time.Time
}

// NewLineWriter returns a new line reader.
func NewLineWriter(stepName, stepUUID string) io.WriteCloser {
	return &LineWriter{
		stepName:  stepName,
		stepUUID:  stepUUID,
		startTime: time.Now().UTC(),
	}
}

func (w *LineWriter) Write(p []byte) (n int, err error) {
	fmt.Fprintf(os.Stderr, "[%s:L%d:%ds] %s", w.stepName, w.num, int64(time.Since(w.startTime).Seconds()), p)
	w.num++
	return len(p), nil
}

func (w *LineWriter) Close() error {
	return nil
}
