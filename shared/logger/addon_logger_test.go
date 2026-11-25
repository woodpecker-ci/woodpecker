// Copyright 2026 Woodpecker Authors
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

package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddonClientLoggerFields(t *testing.T) {
	tests := []struct {
		name string
		log  func(l hclog.Logger)
		want map[string]any
	}{
		{
			name: "key value pairs",
			log: func(l hclog.Logger) {
				l.Error("plugin init error", "error", "open /tmp/plugin123: no such file or directory")
			},
			want: map[string]any{
				"level":   "error",
				"message": "plugin init error",
				"error":   "open /tmp/plugin123: no such file or directory",
			},
		},
		{
			name: "implied args from With",
			log: func(l hclog.Logger) {
				l.With("plugin", "forge").Info("started", "pid", 42)
			},
			want: map[string]any{
				"level":   "info",
				"message": "started",
				"plugin":  "forge",
				"pid":     float64(42),
			},
		},
		{
			name: "value without key",
			log: func(l hclog.Logger) {
				l.Warn("odd args", "key", "value", "dangling")
			},
			want: map[string]any{
				"level":          "warn",
				"message":        "odd args",
				"key":            "value",
				hclog.MissingKey: "dangling",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tt.log(&AddonClientLogger{Logger: zerolog.New(&buf)})

			got := make(map[string]any)
			require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
			assert.Equal(t, tt.want, got)
		})
	}
}
