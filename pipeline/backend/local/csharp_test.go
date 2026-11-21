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
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenCSharpScript(t *testing.T) {
	e := local{tempDir: t.TempDir()}

	tests := []struct {
		name     string
		commands []string
		expected string
	}{
		{
			name: "commands without header",
			commands: []string{
				`Console.WriteLine("first");`,
				`Console.WriteLine("second");`,
			},
			expected: `System.Console.WriteLine(@"▶  Console.WriteLine(""first"");");
Console.WriteLine("first");
System.Console.WriteLine(@"▶  Console.WriteLine(""second"");");
Console.WriteLine("second");
`,
		},
		{
			name: "directives and imports",
			commands: []string{
				`#r "System.Xml"`,
				"using System;",
				"using System.Xml;",
				`Console.WriteLine("hello");`,
			},
			expected: `#r "System.Xml"
using System;
using System.Xml;
System.Console.WriteLine(@"▶  #r ""System.Xml""");
System.Console.WriteLine(@"▶  using System;");
System.Console.WriteLine(@"▶  using System.Xml;");
System.Console.WriteLine(@"▶  Console.WriteLine(""hello"");");
Console.WriteLine("hello");
`,
		},
		{
			name: "comments and multiline commands",
			commands: []string{
				"  // imports\nusing System;",
				"/* header comment */",
				"Console.WriteLine(\"first\");\nConsole.WriteLine(\"second\");",
			},
			expected: `  // imports
using System;
/* header comment */
System.Console.WriteLine(@"▶    // imports
using System;");
System.Console.WriteLine(@"▶  /* header comment */");
System.Console.WriteLine(@"▶  Console.WriteLine(""first"");
Console.WriteLine(""second"");");
Console.WriteLine("first");
Console.WriteLine("second");
`,
		},
		{
			name: "header only",
			commands: []string{
				`#r "System.Xml"`,
				"using System;",
				"// comment",
			},
			expected: `#r "System.Xml"
using System;
// comment
System.Console.WriteLine(@"▶  #r ""System.Xml""");
System.Console.WriteLine(@"▶  using System;");
System.Console.WriteLine(@"▶  // comment");
`,
		},
		{
			name: "using statements are executable",
			commands: []string{
				"using var stream = new System.IO.MemoryStream();",
				"using (var other = new System.IO.MemoryStream()) { }",
			},
			expected: `System.Console.WriteLine(@"▶  using var stream = new System.IO.MemoryStream();");
using var stream = new System.IO.MemoryStream();
System.Console.WriteLine(@"▶  using (var other = new System.IO.MemoryStream()) { }");
using (var other = new System.IO.MemoryStream()) { }
`,
		},
		{
			name: "comments after header",
			commands: []string{
				`Console.WriteLine("first");`,
				"// comment\nConsole.WriteLine(\"second\");",
				"/* comment */\nConsole.WriteLine(\"third\");",
			},
			expected: `System.Console.WriteLine(@"▶  Console.WriteLine(""first"");");
Console.WriteLine("first");
System.Console.WriteLine(@"▶  // comment
Console.WriteLine(""second"");");
// comment
Console.WriteLine("second");
System.Console.WriteLine(@"▶  /* comment */
Console.WriteLine(""third"");");
/* comment */
Console.WriteLine("third");
`,
		},
		{
			name: "backslash escaping",
			commands: []string{
				`System.Console.WriteLine(@"C:\temp\file.txt");`,
			},
			expected: `System.Console.WriteLine(@"▶  System.Console.WriteLine(@""C:\temp\file.txt"");");
System.Console.WriteLine(@"C:\temp\file.txt");
`,
		},
		{
			name: "empty entries in header",
			commands: []string{
				`#r "System.Xml"`,
				"",
				" \t",
				"using System;",
				`Console.WriteLine("hello");`,
			},
			expected: strings.Join([]string{
				`#r "System.Xml"`,
				"",
				" \t",
				"using System;",
				`System.Console.WriteLine(@"▶  #r ""System.Xml""");`,
				`System.Console.WriteLine(@"▶  ");`,
				"System.Console.WriteLine(@\"▶   \t\");",
				`System.Console.WriteLine(@"▶  using System;");`,
				`System.Console.WriteLine(@"▶  Console.WriteLine(""hello"");");`,
				`Console.WriteLine("hello");`,
				"",
			}, "\n"),
		},
		{
			name: "empty entries after header",
			commands: []string{
				`Console.WriteLine("first");`,
				"",
				" ",
				`Console.WriteLine("second");`,
			},
			expected: strings.Join([]string{
				`System.Console.WriteLine(@"▶  Console.WriteLine(""first"");");`,
				`Console.WriteLine("first");`,
				`System.Console.WriteLine(@"▶  ");`,
				"",
				`System.Console.WriteLine(@"▶   ");`,
				" ",
				`System.Console.WriteLine(@"▶  Console.WriteLine(""second"");");`,
				`Console.WriteLine("second");`,
				"",
			}, "\n"),
		},
		{
			name: "only empty entries",
			commands: []string{
				"",
				" \t",
			},
			expected: strings.Join([]string{
				"",
				" \t",
				`System.Console.WriteLine(@"▶  ");`,
				"System.Console.WriteLine(@\"▶   \t\");",
				"",
			}, "\n"),
		},
	}

	for _, shell := range []string{"csi", "dotnet-script", "csharp"} {
		t.Run(shell, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					args, err := e.genCmdByShell(shell, shell, tt.commands, t.TempDir())
					require.NoError(t, err)
					require.Len(t, args, 1)
					assert.True(t, strings.HasSuffix(args[0], ".csx"))

					content, err := os.ReadFile(args[0])
					require.NoError(t, err)
					assert.Equal(t, tt.expected, string(content))
				})
			}
		})
	}
}

func TestCSharpRun(t *testing.T) {
	for _, shell := range []string{"csi", "dotnet-script", "csharp"} {
		t.Run(shell, func(t *testing.T) {
			shellPath, err := exec.LookPath(shell)
			if err != nil {
				t.Skipf("%s not found in PATH", shell)
			}

			e := local{tempDir: t.TempDir()}
			args, err := e.genCmdByShell(shell, shellPath, []string{
				"using System.Globalization;",
				"",
				`System.Console.WriteLine("first");`,
				`System.Console.WriteLine("second");`,
			}, t.TempDir())
			require.NoError(t, err)

			out, err := exec.Command(shellPath, args...).CombinedOutput()
			require.NoError(t, err, string(out))

			output := strings.ReplaceAll(string(out), "\r\n", "\n")
			assert.Equal(t, strings.Join([]string{
				"▶  using System.Globalization;",
				"▶  ",
				`▶  System.Console.WriteLine("first");`,
				"first",
				`▶  System.Console.WriteLine("second");`,
				"second",
				"",
			}, "\n"), output)
		})
	}
}
