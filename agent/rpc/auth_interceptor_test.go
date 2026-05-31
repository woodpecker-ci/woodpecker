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

package rpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestAuthInterceptorAttachToken(t *testing.T) {
	tc := []struct {
		name  string
		token string
	}{
		{"populated token", "secret-token"},
		{"empty token", ""},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			interceptor := &AuthInterceptor{accessToken: c.token}
			ctx := interceptor.attachToken(context.Background())

			md, ok := metadata.FromOutgoingContext(ctx)
			assert.True(t, ok)
			assert.Equal(t, []string{c.token}, md.Get("token"))
		})
	}

	t.Run("preserves existing metadata", func(t *testing.T) {
		base := metadata.AppendToOutgoingContext(context.Background(), "extra", "v")
		interceptor := &AuthInterceptor{accessToken: "tok"}
		ctx := interceptor.attachToken(base)

		md, _ := metadata.FromOutgoingContext(ctx)
		assert.Equal(t, []string{"tok"}, md.Get("token"))
		assert.Equal(t, []string{"v"}, md.Get("extra"))
	})
}
