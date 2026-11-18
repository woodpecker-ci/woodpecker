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

package bitbucketdatacenter

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestWithHTTPClientSkipVerify(t *testing.T) {
	base := context.Background()

	c := &client{skipVerify: false}
	assert.Nil(t, c.withHTTPClient(base).Value(oauth2.HTTPClient))

	c = &client{skipVerify: true}
	hc, ok := c.withHTTPClient(base).Value(oauth2.HTTPClient).(*http.Client)
	require.True(t, ok)
	tr, ok := hc.Transport.(*http.Transport)
	require.True(t, ok)
	assert.True(t, tr.TLSClientConfig.InsecureSkipVerify)
}
