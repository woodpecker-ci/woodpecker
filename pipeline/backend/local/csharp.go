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

package local

import (
	"fmt"
	"os"
	"strings"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/common"
)

func isCSharpHeader(cmd string) bool {
	cmd = strings.TrimSpace(cmd)

	if cmd == "" {
		return true
	}

	if strings.HasPrefix(cmd, "#") ||
		strings.HasPrefix(cmd, "//") ||
		strings.HasPrefix(cmd, "/*") {
		return true
	}

	fields := strings.Fields(cmd)
	if len(fields) < 2 { //nolint:mnd // Command requires at least two fields.
		return false
	}

	if fields[0] == "global" && fields[1] == "using" {
		return true
	}

	if fields[0] != "using" {
		return false
	}

	return fields[1] != "var" && !strings.HasPrefix(fields[1], "(")
}

func genCSharpScript(cmdList []string, baseDir string) ([]string, error) {
	var source strings.Builder
	var deferredMarkers strings.Builder
	header := true

	writeMarker := func(target *strings.Builder, cmd string) {
		fmt.Fprintf(target, "System.Console.WriteLine(@\"%s\");\n",
			strings.ReplaceAll(common.CommandMarker+cmd, `"`, `""`))
	}

	for _, cmd := range cmdList {
		if header && isCSharpHeader(cmd) {
			source.WriteString(cmd + "\n")
			writeMarker(&deferredMarkers, cmd)
			continue
		}

		if header {
			source.WriteString(deferredMarkers.String())
			header = false
		}

		writeMarker(&source, cmd)
		source.WriteString(cmd + "\n")
	}

	if header {
		source.WriteString(deferredMarkers.String())
	}

	script, err := os.CreateTemp(baseDir, "*.csx")
	if err != nil {
		return nil, err
	}
	defer script.Close()

	if _, err := script.WriteString(source.String()); err != nil {
		return nil, err
	}

	return []string{script.Name()}, nil
}
