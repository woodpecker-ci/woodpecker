// Copyright 2023 Woodpecker Authors
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

package grpc

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWTManager is a JSON web token manager
type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

// UserClaims is a custom JWT claims that contains some user's information
type AgentTokenClaims struct {
	jwt.RegisteredClaims
	AgentID int64 `json:"agent_id"`
}

const jwtTokenDuration = 1 * time.Hour

// NewJWTManager returns a new JWT manager
func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{secretKey, jwtTokenDuration}
}

// Generate generates and signs a new token for a user
func (manager *JWTManager) Generate(agentID int64) (string, error) {
	claims := AgentTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "woodpecker",
			Subject:   fmt.Sprintf("%d", agentID),
			Audience:  jwt.ClaimStrings{},
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        fmt.Sprintf("%d", agentID),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(manager.tokenDuration)),
		},
		AgentID: agentID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(manager.secretKey))
}

// Verify verifies the access token string and return a user claim if the token is valid
func (manager *JWTManager) Verify(accessToken string) (*AgentTokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&AgentTokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, errors.New("unexpected token signing method")
			}

			return []byte(manager.secretKey), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*AgentTokenClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
