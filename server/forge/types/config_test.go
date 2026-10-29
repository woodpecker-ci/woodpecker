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

package types

import "testing"

func TestConfigFromURL(t *testing.T) {
	for _, tc := range []struct {
		name string
		url  string
		host string
	}{
		{"root", "https://git.example.com", "git.example.com"},
		{"subdirectory", "https://git.example.com/forgejo/", "git.example.com"},
		{"port and subdirectory", "https://git.example.com:8443/git/", "git.example.com:8443"},
		{"IPv6", "https://[2001:db8::1]:8443/git/", "[2001:db8::1]:8443"},
		{"userinfo", "https://user:password@git.example.com/git/", "git.example.com"},
		{"empty", "", ""},
		{"invalid", "https://git.example.com/%zz", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ConfigFromURL(tc.url).Host; got != tc.host {
				t.Errorf("Host = %q, want %q", got, tc.host)
			}
		})
	}
}
