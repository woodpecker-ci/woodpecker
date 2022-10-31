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

package docker

import (
	"reflect"
	"testing"
)

func TestSplitVolumeParts(t *testing.T) {
	testdata := []struct {
		from    string
		to      []string
		success bool
	}{
		{
			from:    `Z::Z::rw`,
			to:      []string{`Z:`, `Z:`, `rw`},
			success: true,
		},
		{
			from:    `Z:\:Z:\:rw`,
			to:      []string{`Z:\`, `Z:\`, `rw`},
			success: true,
		},
		{
			from:    `Z:\git\refs:Z:\git\refs:rw`,
			to:      []string{`Z:\git\refs`, `Z:\git\refs`, `rw`},
			success: true,
		},
		{
			from:    `Z:\git\refs:Z:\git\refs`,
			to:      []string{`Z:\git\refs`, `Z:\git\refs`},
			success: true,
		},
		{
			from:    `Z:/:Z:/:rw`,
			to:      []string{`Z:/`, `Z:/`, `rw`},
			success: true,
		},
		{
			from:    `Z:/git/refs:Z:/git/refs:rw`,
			to:      []string{`Z:/git/refs`, `Z:/git/refs`, `rw`},
			success: true,
		},
		{
			from:    `Z:/git/refs:Z:/git/refs`,
			to:      []string{`Z:/git/refs`, `Z:/git/refs`},
			success: true,
		},
		{
			from:    `/test:/test`,
			to:      []string{`/test`, `/test`},
			success: true,
		},
		{
			from:    `test:/test`,
			to:      []string{`test`, `/test`},
			success: true,
		},
		{
			from:    `test:test`,
			to:      []string{`test`, `test`},
			success: true,
		},
	}
	for _, test := range testdata {
		results, err := splitVolumeParts(test.from)
		if test.success == (err != nil) {
		} else {
			if reflect.DeepEqual(results, test.to) != test.success {
				t.Errorf("Expect %q matches %q is %v", test.from, results, test.to)
			}
		}
	}
}
