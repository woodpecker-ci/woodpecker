package repo

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

func TestRepoOutput(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
		wantErr  bool
	}{
		{
			name:     "table output with default columns",
			args:     []string{},
			expected: "FULL NAME  BRANCH  FORGE URL        VISIBILITY  SCM PRIVATE  ACTIVE  ALLOW PULL\norg/repo1  main    git.example.com  public      no           yes     yes\n",
		},
		{
			name:     "table output with custom columns",
			args:     []string{"output", "--output", "table=Name,Forge_URL,Trusted_Network"},
			expected: "NAME   FORGE URL        TRUSTED NETWORK\nrepo1  git.example.com  yes\n",
		},
		{
			name:     "table output with no header",
			args:     []string{"output", "--output-no-headers"},
			expected: "org/repo1  main  git.example.com  public  no  yes  yes\n",
		},
		{
			name:     "go-template output",
			args:     []string{"output", "--output", "go-template={{range . }}{{.Name}} {{.ForgeURL}} {{.Trusted.Network}}{{end}}"},
			expected: "repo1 git.example.com true\n",
		},
	}

	repos := []*woodpecker.Repo{
		{
			Name:       "repo1",
			FullName:   "org/repo1",
			ForgeURL:   "git.example.com",
			Branch:     "main",
			Visibility: "public",
			IsActive:   true,
			AllowPull:  true,
			Trusted: woodpecker.TrustedConfiguration{
				Network: true,
			},
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
					err := repoOutput(c, repos, &buf)

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
