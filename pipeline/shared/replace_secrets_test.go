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

package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSecretsReplacer(t *testing.T) {
	tc := []struct {
		name    string
		log     string
		secrets []string
		expect  string
	}{{
		name:    "dont replace secrets with less than 3 chars",
		log:     "start log\ndone",
		secrets: []string{"", "d", "art"},
		expect:  "start log\ndone",
	}, {
		name:    "single line passwords",
		log:     `this IS secret: password`,
		secrets: []string{"password", " IS "},
		expect:  `this IS secret: ********`,
	}, {
		name:    "secret with one newline",
		log:     "start log\ndone\nnow\nan\nmulti line secret!! ;)",
		secrets: []string{"an\nmulti line secret!!"},
		expect:  "start log\ndone\nnow\nan\n******** ;)",
	}, {
		name:    "secret with multiple lines with no match",
		log:     "start log\ndone\nnow\nan\nmulti line secret!! ;)",
		secrets: []string{"Test\nwith\n\ntwo new lines"},
		expect:  "start log\ndone\nnow\nan\nmulti line secret!! ;)",
	}, {
		name:    "secret with multiple lines with match",
		log:     "start log\ndone\nnow\nan\nmulti line secret!! ;)\nwith\ntwo\n\nnewlines",
		secrets: []string{"an\nmulti line secret!!", "two\n\nnewlines"},
		expect:  "start log\ndone\nnow\nan\n******** ;)\nwith\ntwo\n\n********",
	}, {
		name:    "secret with multiple lines with partial match",
		log:     "start with\ntwo",
		secrets: []string{"an\nmulti line secret!!", "two\n\nnewlines"},
		expect:  "start with\ntwo",
	}, {
		name:    "multiline JSON secret does not over-mask short punctuation lines",
		log:     `[{description,"Run PropEr test suites"},{vsn,"0.12.1"},{registered,[]}]`,
		secrets: []string{"{\n\"foo\":[\n\"bar\"\n]\n}"},
		expect:  `[{description,"Run PropEr test suites"},{vsn,"0.12.1"},{registered,[]}]`,
	}}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			rep := NewSecretsReplacer(c.secrets)
			result := rep.Replace(c.log)
			assert.EqualValues(t, c.expect, result)
		})
	}
}
