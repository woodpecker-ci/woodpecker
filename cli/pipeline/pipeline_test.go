package pipeline

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/common"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
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
			expected: "NUMBER  STATUS   EVENT  BRANCH  MESSAGE            AUTHOR\n1       success  push   main    message multiline  John Doe\n",
		},
		{
			name:     "table output with custom columns",
			args:     []string{"output", "--output", "table=Number,Status,Branch"},
			expected: "NUMBER  STATUS   BRANCH\n1       success  main\n",
		},
		{
			name:     "table output with no header",
			args:     []string{"output", "--output-no-headers"},
			expected: "1  success  push  main  message multiline  John Doe\n",
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

	pipelines := []*woodpecker.Pipeline{
		{
			Number:  1,
			Status:  "success",
			Event:   "push",
			Branch:  "main",
			Message: "message\nmultiline",
			Author:  "John Doe\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := &cli.Command{
				Writer: io.Discard,
				Name:   "output",
				Flags:  common.OutputFlags("table"),
				Action: func(_ context.Context, c *cli.Command) error {
					var buf bytes.Buffer
					err := pipelineOutput(c, pipelines, &buf)

					if tt.wantErr {
						assert.Error(t, err)
						return nil
					}

					assert.NoError(t, err)
					assert.Equal(t, tt.expected, buf.String())

					return nil
				},
			}

			_ = command.Run(t.Context(), tt.args)
		})
	}
}
