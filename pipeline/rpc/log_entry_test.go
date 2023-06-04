// Copyright 2019 Woodpecker Authors
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

package rpc

import (
	"testing"
)

func TestLogEntry(t *testing.T) {
	line := LogEntry{
		StepUUID: "e9ea76a5-44a1-4059-9c4a-6956c478b26d",
		Time:     60,
		Line:     1,
		Data:     "starting redis server",
	}
	got, want := line.String(), "[redis:L1:60s] starting redis server"
	if got != want {
		t.Errorf("Wanted line string %q, got %q", want, got)
	}
}
