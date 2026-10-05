// Copyright 2026 Woodpecker Authors
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

package docker

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

type pullResponse struct {
	client.ImagePullResponse
	io.ReadCloser
}

func (r pullResponse) Read(p []byte) (int, error) { return r.ReadCloser.Read(p) }
func (r pullResponse) Close() error               { return r.ReadCloser.Close() }

type pullClient struct {
	client.APIClient
	stream string
}

func (c pullClient) ImagePull(context.Context, string, client.ImagePullOptions) (client.ImagePullResponse, error) {
	return pullResponse{ReadCloser: io.NopCloser(strings.NewReader(c.stream))}, nil
}

func TestPullImageOutput(t *testing.T) {
	e := &docker{client: pullClient{stream: `{"status":"Pulling from library/alpine","id":"latest"}
{"status":"Downloading","progressDetail":{"current":10,"total":100},"id":"abc"}
{"status":"Pull complete","id":"abc"}
{"status":"Status: Downloaded newer image for alpine:latest"}
`}}

	out := new(bytes.Buffer)
	ctx := context.WithValue(t.Context(), backend_types.ImagePullOutput, out)
	require.NoError(t, e.pullImage(ctx, "alpine:latest", client.ImagePullOptions{}))

	// command heading first, progress bars are left out as it is no terminal
	assert.Equal(t, `▶  docker pull alpine:latest
latest: Pulling from library/alpine
abc: Pull complete
Status: Downloaded newer image for alpine:latest
`, out.String())
}

func TestPullImageStreamError(t *testing.T) {
	e := &docker{client: pullClient{stream: `{"status":"Pulling from library/nope","id":"latest"}
{"errorDetail":{"message":"manifest unknown"},"error":"manifest unknown"}
`}}

	out := new(bytes.Buffer)
	ctx := context.WithValue(t.Context(), backend_types.ImagePullOutput, out)

	// the reason must end up as step error, not only in the agent log
	assert.EqualError(t, e.pullImage(ctx, "nope:latest", client.ImagePullOptions{}), "manifest unknown")
	assert.Equal(t, "▶  docker pull nope:latest\nlatest: Pulling from library/nope\n", out.String())
}
