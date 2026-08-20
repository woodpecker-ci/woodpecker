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
	"crypto/subtle"
	"fmt"

	"go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

// TokenType tells for which feature a repo token grants access.
type TokenType string //	@name	TokenType

const (
	// TokenTypeBadge grants read access to the badge endpoints of a repo.
	TokenTypeBadge TokenType = "badge"
)

// tokenLength is the number of characters a generated token holds.
const tokenLength = 52

// Valid reports whether the token type is a known one.
func (t TokenType) Valid() bool {
	switch t {
	case TokenTypeBadge:
		return true
	default:
		return false
	}
}

// Token is a secret bound to a repo, granting access to a single feature of it.
type Token struct {
	ID      int64     `json:"id"      xorm:"pk autoincr 'id'"`
	RepoID  int64     `json:"repo_id" xorm:"NOT NULL UNIQUE(s) INDEX 'repo_id'"`
	Type    TokenType `json:"type"    xorm:"NOT NULL UNIQUE(s) INDEX varchar(50) 'type'"`
	Value   string    `json:"value"   xorm:"NOT NULL UNIQUE varchar(100) 'value'"`
	Created int64     `json:"created" xorm:"created NOT NULL DEFAULT 0"`
} //	@name	Token

// TableName returns the database table name for xorm.
func (Token) TableName() string {
	return "tokens"
}

// Validate ensures the token is bound to a repo and has a known type and a value.
func (t *Token) Validate() error {
	if t.RepoID == 0 {
		return fmt.Errorf("repo id is required")
	}

	if !t.Type.Valid() {
		return fmt.Errorf("invalid token type '%s'", t.Type)
	}

	if t.Value == "" {
		return fmt.Errorf("value is required")
	}

	return nil
}

// Matches compares the token value against the given one in constant time.
func (t *Token) Matches(value string) bool {
	return subtle.ConstantTimeCompare([]byte(t.Value), []byte(value)) == 1
}

// NewToken returns a token of the given type with a freshly generated value,
// bound to the given repo.
func NewToken(repoID int64, tokenType TokenType) *Token {
	return &Token{
		RepoID: repoID,
		Type:   tokenType,
		Value:  GenerateTokenValue(),
	}
}

// GenerateTokenValue returns a new random token value.
func GenerateTokenValue() string {
	return utils.RandomString(tokenLength)
}
