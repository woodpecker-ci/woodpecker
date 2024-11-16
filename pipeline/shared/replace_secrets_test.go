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
	"bytes"
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
		name:    "dont replace secrets with less than 4 chars",
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
		expect:  "start log\ndone\nnow\n********\n******** ;)",
	}, {
		name:    "secret with multiple lines with no match",
		log:     "start log\ndone\nnow\nan\nmulti line secret!! ;)",
		secrets: []string{"Test\nwith\n\ntwo new lines"},
		expect:  "start log\ndone\nnow\nan\nmulti line secret!! ;)",
	}, {
		name:    "secret with multiple lines with match",
		log:     "start log\ndone\nnow\nan\nmulti line secret!! ;)\nwith\ntwo\n\nnewlines",
		secrets: []string{"an\nmulti line secret!!", "two\n\nnewlines"},
		expect:  "start log\ndone\nnow\n********\n******** ;)\nwith\n********\n\n********",
	}}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			rep := NewSecretsReplacer(c.secrets)
			result := rep.Replace(c.log)
			assert.EqualValues(t, c.expect, result)
		})
	}
}

func BenchmarkReader(b *testing.B) {
	testCases := []struct {
		name    string
		log     string
		secrets []string
	}{
		{
			name:    "single line",
			log:     "this is a log with secret password and more text",
			secrets: []string{"password"},
		},
		{
			name:    "multi line",
			log:     "log start\nthis is a multi\nline secret\nlog end",
			secrets: []string{"multi\nline secret"},
		},
		{
			name:    "many secrets",
			log:     "log with many secrets: secret1 secret2 secret3 secret4 secret5",
			secrets: []string{"secret1", "secret2", "secret3", "secret4", "secret5"},
		},
		{
			name:    "large log",
			log:     "start " + string(bytes.Repeat([]byte("test secret test "), 1000)) + " end",
			secrets: []string{"secret"},
		},
		{
			name:    "large log no match",
			log:     "start " + string(bytes.Repeat([]byte("test secret test "), 1000)) + " end",
			secrets: []string{"XXXXXXX"},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			rep := NewSecretsReplacer(tc.secrets)
			b.ResetTimer()
			b.SetBytes(int64(len(tc.log)))
			for i := 0; i < b.N; i++ {
				_ = rep.Replace(tc.log)
			}
		})
	}
}

// go test -benchmem -run='^$' -tags test -bench '^BenchmarkReader$' -benchtime=100000x go.woodpecker-ci.org/woodpecker/v2/pipeline/shared
//
// cpu: AMD Ryzen 9 7940HS
// BenchmarkReader/single_line-16            100000             55.13 ns/op         870.70 MB/s          48 B/op          1 allocs/op
// BenchmarkReader/multi_line-16             100000             149.0 ns/op         302.06 MB/s         120 B/op          3 allocs/op
// BenchmarkReader/many_secrets-16           100000             273.0 ns/op         227.10 MB/s         296 B/op          4 allocs/op
// BenchmarkReader/large_log-16              100000             19544 ns/op         870.33 MB/s       40520 B/op          9 allocs/op
// BenchmarkReader/large_log_no_match-16     100000              5080 ns/op        3348.63 MB/s           0 B/op          0 allocs/op
//
// cpu: AMD Ryzen 9 3900XT 12-Core Processor
// BenchmarkReader/single_line-24            100000                90.87 ns/op      528.23 MB/s          48 B/op          1 allocs/op
// BenchmarkReader/multi_line-24             100000               276.2 ns/op       162.94 MB/s         120 B/op          3 allocs/op
// BenchmarkReader/many_secrets-24           100000               433.7 ns/op       142.97 MB/s         296 B/op          4 allocs/op
// BenchmarkReader/large_log-24              100000             26542 ns/op         640.88 MB/s       40520 B/op          9 allocs/op
// BenchmarkReader/large_log_no_match-24     100000              6212 ns/op        2738.45 MB/s           0 B/op          0 allocs/op
//
// cpu: Ampere Altra 2 vCPUs
// BenchmarkReader/single_line-2             100000             105.1 ns/op         456.89 MB/s          48 B/op          1 allocs/op
// BenchmarkReader/multi_line-2              100000             441.7 ns/op         101.88 MB/s         120 B/op          3 allocs/op
// BenchmarkReader/many_secrets-2            100000             868.7 ns/op          71.37 MB/s         296 B/op          4 allocs/op
// BenchmarkReader/large_log-2               100000             48947 ns/op         347.52 MB/s       40520 B/op          9 allocs/op
// BenchmarkReader/large_log_no_match-2      100000              9156 ns/op        1857.79 MB/s           0 B/op          0 allocs/op

// TODO: benchmark arm & intel
