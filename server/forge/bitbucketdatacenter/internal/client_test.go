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

package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestCurrentUser(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`tal@netic.dk`))
	}))

	defer s.Close()

	ctx := context.Background()
	ts := mockSource("bearer-token")
	client := NewClientWithToken(ctx, ts, s.URL)
	uid, err := client.FindCurrentUser(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "tal_netic.dk", uid)
}

type mockSource string

func (ds mockSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: string(ds)}, nil
}
