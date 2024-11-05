// Copyright 2024 Woodpecker Authors
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

package exec

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/matrix"
)

func TestMetadataFromContext(t *testing.T) {
	sampleMetadata := &metadata.Metadata{
		Repo: metadata.Repo{Owner: "test-user", Name: "test-repo"},
		Curr: metadata.Pipeline{Number: 5},
	}

	runCommand := func(flags []cli.Flag, fn func(c *cli.Command)) {
		c := &cli.Command{
			Flags: flags,
			Action: func(_ context.Context, c *cli.Command) error {
				fn(c)
				return nil
			},
		}
		assert.NoError(t, c.Run(context.Background(), []string{"woodpecker-cli"}))
	}

	t.Run("LoadFromFile", func(t *testing.T) {
		tempFileName := createTempFile(t, sampleMetadata)

		flags := []cli.Flag{
			&cli.StringFlag{Name: "metadata-file"},
		}

		runCommand(flags, func(c *cli.Command) {
			_ = c.Set("metadata-file", tempFileName)

			m, err := metadataFromContext(context.Background(), c, nil, nil)
			require.NoError(t, err)
			assert.Equal(t, "test-repo", m.Repo.Name)
			assert.Equal(t, int64(5), m.Curr.Number)
		})
	})

	t.Run("OverrideFromFlags", func(t *testing.T) {
		tempFileName := createTempFile(t, sampleMetadata)

		flags := []cli.Flag{
			&cli.StringFlag{Name: "metadata-file"},
			&cli.StringFlag{Name: "repo-name"},
			&cli.IntFlag{Name: "pipeline-number"},
		}

		runCommand(flags, func(c *cli.Command) {
			_ = c.Set("metadata-file", tempFileName)
			_ = c.Set("repo-name", "aUser/override-repo")
			_ = c.Set("pipeline-number", "10")

			m, err := metadataFromContext(context.Background(), c, nil, nil)
			require.NoError(t, err)
			assert.Equal(t, "override-repo", m.Repo.Name)
			assert.Equal(t, int64(10), m.Curr.Number)
		})
	})

	t.Run("InvalidFile", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "invalid.json")
		require.NoError(t, err)
		t.Cleanup(func() { os.Remove(tempFile.Name()) })

		_, err = tempFile.Write([]byte("invalid json"))
		require.NoError(t, err)

		flags := []cli.Flag{
			&cli.StringFlag{Name: "metadata-file"},
		}

		runCommand(flags, func(c *cli.Command) {
			_ = c.Set("metadata-file", tempFile.Name())

			_, err = metadataFromContext(context.Background(), c, nil, nil)
			assert.Error(t, err)
		})
	})

	t.Run("DefaultValues", func(t *testing.T) {
		flags := []cli.Flag{
			&cli.StringFlag{Name: "repo-name", Value: "test/default-repo"},
			&cli.IntFlag{Name: "pipeline-number", Value: 1},
		}

		runCommand(flags, func(c *cli.Command) {
			m, err := metadataFromContext(context.Background(), c, nil, nil)
			require.NoError(t, err)
			if assert.NotNil(t, m) {
				assert.Equal(t, "test", m.Repo.Owner)
				assert.Equal(t, "default-repo", m.Repo.Name)
				assert.Equal(t, int64(1), m.Curr.Number)
			}
		})
	})

	t.Run("MatrixAxis", func(t *testing.T) {
		runCommand([]cli.Flag{}, func(c *cli.Command) {
			axis := matrix.Axis{"go": "1.16", "os": "linux"}
			m, err := metadataFromContext(context.Background(), c, axis, nil)
			require.NoError(t, err)
			assert.EqualValues(t, map[string]string{"go": "1.16", "os": "linux"}, m.Workflow.Matrix)
		})
	})
}

func createTempFile(t *testing.T, content any) string {
	t.Helper()
	tempFile, err := os.CreateTemp("", "metadata.json")
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(tempFile.Name()) })

	err = json.NewEncoder(tempFile).Encode(content)
	require.NoError(t, err)
	return tempFile.Name()
}
