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

// Guest is a minimal configuration extension to test the wasm transport.
//
// It converts ".steps" files to workflows: every line becomes a step of that name.
// The placeholders "{event}" and "{repo}" are replaced with values of the request,
// a line starting with "!fail " aborts the conversion with the rest of the line as message
// and a line starting with "!raw " makes the rest of the line the whole answer.
// Without any ".steps" file it answers nothing, so the server keeps the configuration it has.
package main

import (
	"encoding/json"
	"os"
	"strings"
)

type file struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type request struct {
	Repo struct {
		Name string `json:"name"`
	} `json:"repo"`
	Pipeline struct {
		Event string `json:"event"`
	} `json:"pipeline"`
	Netrc         any    `json:"netrc"`
	Configuration []file `json:"configuration"`
}

const ext = ".steps"

func main() {
	if len(os.Args) != 2 || os.Args[1] != "config" { //nolint:mnd
		fail("unexpected arguments: " + strings.Join(os.Args, " "))
	}

	var req request
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		fail("invalid request: " + err.Error())
	}
	if req.Netrc != nil {
		fail("got netrc data")
	}

	var configs []file
	converted := false
	for _, f := range req.Configuration {
		name, ok := strings.CutSuffix(f.Name, ext)
		if !ok {
			configs = append(configs, f)
			continue
		}
		converted = true

		data := strings.NewReplacer("{event}", req.Pipeline.Event, "{repo}", req.Repo.Name).Replace(f.Data)
		var yaml strings.Builder
		yaml.WriteString("steps:\n")
		for line := range strings.Lines(data) {
			line = strings.TrimSpace(line)
			if msg, ok := strings.CutPrefix(line, "!fail "); ok {
				fail(msg)
			}
			if raw, ok := strings.CutPrefix(line, "!raw "); ok {
				_, _ = os.Stdout.WriteString(raw)
				return
			}
			if line != "" {
				yaml.WriteString("  - name: " + line + "\n    image: dummy\n    commands:\n      - echo " + line + "\n")
			}
		}
		configs = append(configs, file{Name: name + ".yaml", Data: yaml.String()})
	}

	if !converted {
		return
	}

	if err := json.NewEncoder(os.Stdout).Encode(map[string]any{"configs": configs}); err != nil {
		fail(err.Error())
	}
}

func fail(msg string) {
	_, _ = os.Stderr.WriteString(msg + "\n")
	os.Exit(1)
}
