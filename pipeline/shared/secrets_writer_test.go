// Copyright 2024 Woodpecker Authors
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
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecretsReplaceWriter(t *testing.T) {
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
		name:    "secret multiple times",
		log:     "token token2 token\n",
		secrets: []string{"token"},
		expect:  "******** ********2 ********\n",
	}, {
		name:    "multiple secrets in one line",
		log:     "user=admin1 pass=hunter2\n",
		secrets: []string{"hunter2", "admin1"},
		expect:  "user=******** pass=********\n",
	}, {
		name:    "longer secret wins over a shorter one it contains",
		log:     "my password and my pass\n",
		secrets: []string{"pass", "password"},
		expect:  "my ******** and my ********\n",
	}, {
		name:    "longer secret wins over an earlier overlapping one",
		log:     "abcde\n",
		secrets: []string{"abcd", "bcdef", "cde"},
		expect:  "********e\n",
	}, {
		name:    "longer secret wins even if it starts later",
		log:     "xabcdefx\n",
		secrets: []string{"xabc", "bcdef"},
		expect:  "xa********x\n",
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
			var buf bytes.Buffer
			w := NewSecretsReplaceWriter(&buf, c.secrets)

			n, err := w.Write([]byte(c.log))
			assert.NoError(t, err)
			assert.Equal(t, len(c.log), n)
			assert.Equal(t, c.expect, buf.String())
		})
	}
}

// TestSecretsReplaceWriterKeepsInput guards the io.Writer contract: masking
// must not modify the caller's buffer.
func TestSecretsReplaceWriterKeepsInput(t *testing.T) {
	var buf bytes.Buffer
	w := NewSecretsReplaceWriter(&buf, []string{"supersecret"})

	in := []byte("token is supersecret\n")
	_, err := w.Write(in)
	assert.NoError(t, err)
	assert.Equal(t, "token is supersecret\n", string(in))
	assert.Equal(t, "token is ********\n", buf.String())
}

func TestSecretsReplaceWriterNoSecrets(t *testing.T) {
	var buf bytes.Buffer
	w := NewSecretsReplaceWriter(&buf, nil)

	_, err := w.Write([]byte("plain\n"))
	assert.NoError(t, err)
	assert.Equal(t, "plain\n", buf.String())
}

type errWriter struct{ err error }

func (e *errWriter) Write([]byte) (int, error) { return 0, e.err }

func TestSecretsReplaceWriterPropagatesError(t *testing.T) {
	wantErr := errors.New("sink closed")
	w := NewSecretsReplaceWriter(&errWriter{err: wantErr}, nil)

	n, err := w.Write([]byte("x"))
	assert.ErrorIs(t, err, wantErr)
	assert.Zero(t, n)
}
