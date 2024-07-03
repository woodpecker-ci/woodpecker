package pipeline

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

func TestPipelineOutput(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
		wantErr  bool
	}{
		{
			name:     "table output with default columns",
			args:     []string{},
			expected: "NUMBER  STATUS   EVENT  BRANCH  MESSAGE  AUTHOR\n1       success  push   main    message  John Doe\n",
		},
		{
			name:     "table output with custom columns",
			args:     []string{"output", "--output", "table=Number,Status,Branch"},
			expected: "NUMBER  STATUS   BRANCH\n1       success  main\n",
		},
		{
			name:     "table output with no header",
			args:     []string{"output", "--output-no-headers"},
			expected: "1  success  push  main  message  John Doe\n",
		},
		{
			name:     "go-template output",
			args:     []string{"output", "--output", "go-template={{range . }}{{.Number}} {{.Status}} {{.Branch}}{{end}}"},
			expected: "1 success main\n",
		},
		{
			name:    "invalid go-template",
			args:    []string{"output", "--output", "go-template={{.InvalidField}}"},
			wantErr: true,
		},
	}

	pipelines := []woodpecker.Pipeline{
		{
			Number:  1,
			Status:  "success",
			Event:   "push",
			Branch:  "main",
			Message: "message",
			Author:  "John Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &cli.App{Writer: io.Discard}
			c := cli.NewContext(app, nil, nil)

			command := &cli.Command{}
			command.Name = "output"
			command.Flags = common.OutputFlags("table")
			command.Action = func(c *cli.Context) error {
				var buf bytes.Buffer
				err := pipelineOutput(c, pipelines, &buf)

				if tt.wantErr {
					assert.Error(t, err)
					return nil
				}

				assert.NoError(t, err)
				assert.Equal(t, tt.expected, buf.String())

				return nil
			}

			_ = command.Run(c, tt.args...)
		})
	}
}
