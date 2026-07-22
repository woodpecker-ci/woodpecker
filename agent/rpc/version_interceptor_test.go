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

package rpc

import (
	"sync/atomic"
	"testing"

	"google.golang.org/grpc/metadata"

	"go.woodpecker-ci.org/woodpecker/v3/rpc"
)

func mdWith(serverVersion string) metadata.MD {
	if serverVersion == "" {
		return metadata.MD{}
	}
	return metadata.Pairs(rpc.MetadataKeyServerVersion, serverVersion)
}

// TestVersionInterceptorCheck encodes the per-method policy maintainers
// asked for: shutdown on Next/Init, ignore on Version, warn (no shutdown)
// on everything else, and skip checks entirely when either side is "dev".
func TestVersionInterceptorCheck(t *testing.T) {
	cases := []struct {
		name           string
		agentVersion   string
		method         string
		md             metadata.MD
		wantShutdownOn string // server version expected at shutdown handler, "" = no call
	}{
		{"agent dev disables check", "dev", "/proto.Woodpecker/Next", mdWith("v3.15.0"), ""},
		{"server dev skipped", "v3.14.0", "/proto.Woodpecker/Next", mdWith("dev"), ""},
		{"matching versions", "v3.14.0", "/proto.Woodpecker/Next", mdWith("v3.14.0"), ""},
		{"missing header", "v3.14.0", "/proto.Woodpecker/Next", mdWith(""), ""},
		{"version method ignored", "v3.14.0", "/proto.Woodpecker/Version", mdWith("v3.15.0"), ""},
		{"extend warns only", "v3.14.0", "/proto.Woodpecker/Extend", mdWith("v3.15.0"), ""},
		{"done warns only", "v3.14.0", "/proto.Woodpecker/Done", mdWith("v3.15.0"), ""},
		{"register agent warns only", "v3.14.0", "/proto.Woodpecker/RegisterAgent", mdWith("v3.15.0"), ""},
		{"next triggers shutdown", "v3.14.0", "/proto.Woodpecker/Next", mdWith("v3.15.0"), "v3.15.0"},
		{"init triggers shutdown", "v3.14.0", "/proto.Woodpecker/Init", mdWith("v3.15.0"), "v3.15.0"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vi := NewVersionInterceptor(tc.agentVersion)
			var got atomic.Value
			vi.SetShutdownHandler(func(serverVersion string) {
				got.Store(serverVersion)
			})

			vi.check(tc.method, tc.md)

			v, _ := got.Load().(string)
			if v != tc.wantShutdownOn {
				t.Fatalf("shutdown handler got %q, want %q", v, tc.wantShutdownOn)
			}
		})
	}
}

// TestVersionInterceptorShutdownOnce verifies the shutdown handler fires at
// most once, even when several mismatched RPCs land back-to-back.
func TestVersionInterceptorShutdownOnce(t *testing.T) {
	vi := NewVersionInterceptor("v3.14.0")
	var calls atomic.Int32
	vi.SetShutdownHandler(func(string) {
		calls.Add(1)
	})

	for i := 0; i < 5; i++ {
		vi.check("/proto.Woodpecker/Next", mdWith("v3.15.0"))
	}
	vi.check("/proto.Woodpecker/Init", mdWith("v3.15.0"))

	if got := calls.Load(); got != 1 {
		t.Fatalf("shutdown handler called %d times, want 1", got)
	}
}
