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

// Guest is a WASI module used to test the sandbox of the wasm runner.
// The first argument selects what it does.
package main

import (
	"io"
	"os"
	"strconv"
	"strings"
)

var calls int

func main() {
	calls++

	mode := ""
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	switch mode {
	case "echo":
		_, _ = io.Copy(os.Stdout, os.Stdin)
	case "args":
		_, _ = os.Stdout.WriteString(strings.Join(os.Args, ","))
	case "calls":
		_, _ = os.Stdout.WriteString(strconv.Itoa(calls))
	case "fail":
		_, _ = os.Stderr.WriteString("something went wrong")
		os.Exit(3)
	case "spin":
		for {
		}
	case "alloc":
		var hold [][]byte
		for {
			hold = append(hold, make([]byte, 1<<20))
			hold[len(hold)-1][0] = 1
		}
	case "flood-stdout":
		chunk := make([]byte, 1<<16)
		for {
			if _, err := os.Stdout.Write(chunk); err != nil {
				os.Exit(4)
			}
		}
	case "flood-stderr":
		chunk := []byte(strings.Repeat("e", 1<<16))
		for range 64 {
			_, _ = os.Stderr.Write(chunk)
		}
		_, _ = os.Stdout.WriteString("done")
	case "probe":
		probe()
	default:
		_, _ = os.Stderr.WriteString("unknown mode " + mode)
		os.Exit(2)
	}
}

// probe reports what the guest can reach outside of its sandbox, one finding per line.
// WASI has no call to open a connection, so files are the only way out that can be probed.
func probe() {
	_, rootErr := os.ReadDir("/")
	_, cwdErr := os.ReadDir(".")
	writeErr := os.WriteFile("/tmp/escape", []byte("x"), 0o600)
	// the first descriptor after stdin, stdout and stderr is where a shared directory or socket would be
	extraErr := os.NewFile(3, "fd3").Close() //nolint:mnd

	for _, finding := range []struct {
		name    string
		reached bool
	}{
		{"env", len(os.Environ()) != 0},
		{"read_root", rootErr == nil},
		{"read_cwd", cwdErr == nil},
		{"write_file", writeErr == nil},
		{"extra_descriptor", extraErr == nil},
	} {
		_, _ = os.Stdout.WriteString(finding.name + "=" + strconv.FormatBool(finding.reached) + "\n")
	}
}
