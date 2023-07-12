// Copyright 2023 Woodpecker Authors
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

package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadAgentIDFileNotExists(t *testing.T) {
	assert.EqualValues(t, -1, readAgentID("foobar.conf"))
}

func TestReadAgentIDFileExists(t *testing.T) {
	parameters := []struct {
		input    string
		expected int64
	}{
		{"42", 42},
		{"42\n", 42},
		{"  \t42\t\r\t", 42},
		{"0", 0},
		{"-1", -1},
		{"foo", -1},
		{"1f", -1},
		{"", -1},
		{"-42", -42},
	}

	for i := range parameters {
		t.Run(fmt.Sprintf("Testing [%v]", i), func(t *testing.T) {
			tmpF, errTmpF := os.CreateTemp("", "tmp_")
			if !assert.NoError(t, errTmpF) {
				t.FailNow()
			}

			errWrite := os.WriteFile(tmpF.Name(), []byte(parameters[i].input), 0o644)
			if !assert.NoError(t, errWrite) {
				t.FailNow()
			}

			actual := readAgentID(tmpF.Name())
			assert.EqualValues(t, parameters[i].expected, actual)
		})
	}
}
