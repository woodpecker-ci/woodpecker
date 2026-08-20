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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenValidate(t *testing.T) {
	tests := []struct {
		name    string
		token   Token
		wantErr bool
	}{
		{
			name:  "valid badge token",
			token: Token{RepoID: 1, Type: TokenTypeBadge, Value: "secret"},
		},
		{
			name:    "missing repo",
			token:   Token{Type: TokenTypeBadge, Value: "secret"},
			wantErr: true,
		},
		{
			name:    "negative repo",
			token:   Token{RepoID: -1, Type: TokenTypeBadge, Value: "secret"},
			wantErr: true,
		},
		{
			name:    "unknown type",
			token:   Token{RepoID: 1, Type: TokenType("cron"), Value: "secret"},
			wantErr: true,
		},
		{
			name:    "missing value",
			token:   Token{RepoID: 1, Type: TokenTypeBadge},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.token.Validate()
			if test.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestTokenMatches(t *testing.T) {
	token := Token{RepoID: 1, Type: TokenTypeBadge, Value: "secret"}

	assert.True(t, token.Matches("secret"))
	assert.False(t, token.Matches("secre"))
	assert.False(t, token.Matches("secretsecret"))
	assert.False(t, token.Matches(""))
}

func TestNewToken(t *testing.T) {
	first := NewToken(3, TokenTypeBadge)
	second := NewToken(3, TokenTypeBadge)

	assert.NoError(t, first.Validate())
	assert.Equal(t, int64(3), first.RepoID)
	assert.Equal(t, TokenTypeBadge, first.Type)
	assert.NotEqual(t, first.Value, second.Value)
}
