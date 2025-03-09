// Copyright 2024 Woodpecker Authors
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

package docker

import "testing"

func TestMustNotAddWindowsLetterPattern(t *testing.T) {
	tests := map[string]bool{
		`C:\Users`:           true,
		`D:\Data`:            true,
		`\\.\PhysicalDrive0`: true,
		`//./COM1`:           true,
		`E:`:                 true,
		`\\server\share`:     true, // UNC path
		`.\relative\path`:    true, // Relative path
		`./path`:             true, // Relative with forward slash
		`//server/share`:     true, // UNC with forward slashes
		`not/a/windows/path`: false,
		``:                   false,
		`/usr/local`:         false,
		`COM1`:               false,
		`\\.`:                false, // Incomplete device path
		`//`:                 false,
	}

	for testCase, expected := range tests {
		result := mustNotAddWindowsLetterPattern.MatchString(testCase)
		if result != expected {
			t.Errorf("Test case %q: expected %v but got %v", testCase, expected, result)
		}
	}
}
