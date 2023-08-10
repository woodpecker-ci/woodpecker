// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import "testing"

func TestParseKeyPair(t *testing.T) {
	s := []string{"FOO=bar", "BAR=", "BAZ=qux=quux", "INVALID"}
	p := ParseKeyPair(s)
	if p["FOO"] != "bar" {
		t.Errorf("Wanted %q, got %q.", "bar", p["FOO"])
	}
	if p["BAZ"] != "qux=quux" {
		t.Errorf("Wanted %q, got %q.", "qux=quux", p["BAZ"])
	}
	if _, exists := p["BAR"]; !exists {
		t.Error("Missing a key with no value. Keys with empty values are also valid.")
	}
	if _, exists := p["INVALID"]; exists {
		t.Error("Keys without an equal sign suffix are invalid.")
	}
}
