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
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestStringSliceAddToMap(t *testing.T) {
	tests := []struct {
		name     string
		sl       []string
		m        map[string]string
		expected map[string]string
		err      bool
	}{
		{
			name: "add values to map",
			sl:   []string{"foo=bar", "baz=qux=nux"},
			m:    make(map[string]string),
			expected: map[string]string{
				"foo": "bar",
				"baz": "qux=nux",
			},
			err: false,
		},
		{
			name:     "empty slice",
			sl:       []string{},
			m:        make(map[string]string),
			expected: map[string]string{},
			err:      false,
		},
		{
			name:     "missing value",
			sl:       []string{"foo", "baz=qux"},
			m:        make(map[string]string),
			expected: map[string]string{},
			err:      true,
		},
		{
			name:     "empty string in slice",
			sl:       []string{"foo=bar", "", "baz=qux"},
			m:        make(map[string]string),
			expected: map[string]string{},
			err:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := stringSliceAddToMap(tt.sl, tt.m)

			if tt.err {
				assert.Error(t, err)
			} else {
				assert.EqualValues(t, tt.expected, tt.m)
			}
		})
	}
}

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

func TestWriteAgentIDFileNotExists(t *testing.T) {
	tmpF, errTmpF := os.CreateTemp("", "tmp_")
	if !assert.NoError(t, errTmpF) {
		t.FailNow()
	}

	writeAgentID(42, tmpF.Name())
	actual, errRead := os.ReadFile(tmpF.Name())
	if !assert.NoError(t, errRead) {
		t.FailNow()
	}
	assert.EqualValues(t, "42\n", actual)
}
func TestWriteAgentIDFileExists(t *testing.T) {
	parameters := []struct {
		fileInput  string
		writeInput int64
		expected   string
	}{
		{"", 42, "42\n"},
		{"\n", 42, "42\n"},
		{"41\n", 42, "42\n"},
		{"0", 42, "42\n"},
		{"-1", 42, "42\n"},
		{"foöbar", 42, "42\n"},
	}

	for i := range parameters {
		t.Run(fmt.Sprintf("Testing [%v]", i), func(t *testing.T) {
			tmpF, errTmpF := os.CreateTemp("", "tmp_")
			if !assert.NoError(t, errTmpF) {
				t.FailNow()
			}

			errWrite := os.WriteFile(tmpF.Name(), []byte(parameters[i].fileInput), 0o644)
			if !assert.NoError(t, errWrite) {
				t.FailNow()
			}

			writeAgentID(parameters[i].writeInput, tmpF.Name())
			actual, errRead := os.ReadFile(tmpF.Name())
			if !assert.NoError(t, errRead) {
				t.FailNow()
			}
			assert.EqualValues(t, parameters[i].expected, actual)
		})
	}
}
