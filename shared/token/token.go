// Copyright 2018 Drone.IO Inc.
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

package token

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

type SecretFunc func(*Token) (string, error)

const (
	UserToken  = "user"
	SessToken  = "sess"
	HookToken  = "hook"
	CsrfToken  = "csrf"
	AgentToken = "agent"
)

// SignerAlgo id default algorithm used to sign JWT tokens.
const SignerAlgo = "HS256"

type Token struct {
	Kind string
	Text string
}

func parse(raw string, fn SecretFunc) (*Token, error) {
	token := &Token{}
	parsed, err := jwt.Parse(raw, keyFunc(token, fn))
	if err != nil {
		return nil, err
	} else if !parsed.Valid {
		return nil, jwt.ErrTokenUnverifiable
	}
	return token, nil
}

func ParseRequest(r *http.Request, fn SecretFunc) (*Token, error) {
	// first we attempt to get the token from the
	// authorization header.
	token := r.Header.Get("Authorization")
	if len(token) != 0 {
		log.Trace().Msgf("token.ParseRequest: found token in header: %s", token)
		bearer := token
		if _, err := fmt.Sscanf(token, "Bearer %s", &bearer); err != nil {
			return nil, err
		}
		return parse(bearer, fn)
	}

	token = r.Header.Get("X-Gitlab-Token")
	if len(token) != 0 {
		return parse(token, fn)
	}

	// then we attempt to get the token from the
	// access_token url query parameter
	token = r.FormValue("access_token")
	if len(token) != 0 {
		return parse(token, fn)
	}

	// and finally we attempt to get the token from
	// the user session cookie
	cookie, err := r.Cookie("user_sess")
	if err != nil {
		return nil, err
	}
	return parse(cookie.Value, fn)
}

func CheckCsrf(r *http.Request, fn SecretFunc) error {
	// get and options requests are always
	// enabled, without CSRF checks.
	switch r.Method {
	case "GET", "OPTIONS":
		return nil
	}

	// parse the raw CSRF token value and validate
	raw := r.Header.Get("X-CSRF-TOKEN")
	_, err := parse(raw, fn)
	return err
}

func New(kind, text string) *Token {
	return &Token{Kind: kind, Text: text}
}

// Sign signs the token using the given secret hash
// and returns the string value.
func (t *Token) Sign(secret string) (string, error) {
	return t.SignExpires(secret, 0)
}

// Sign signs the token using the given secret hash
// with an expiration date.
func (t *Token) SignExpires(secret string, exp int64) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("token claim is not a MapClaims")
	}
	claims["type"] = t.Kind
	claims["text"] = t.Text
	if exp > 0 {
		claims["exp"] = float64(exp)
	}
	return token.SignedString([]byte(secret))
}

func keyFunc(token *Token, fn SecretFunc) jwt.Keyfunc {
	return func(t *jwt.Token) (interface{}, error) {
		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("token claim is not a MapClaims")
		}

		// validate the correct algorithm is being used
		if t.Method.Alg() != SignerAlgo {
			return nil, jwt.ErrSignatureInvalid
		}

		// extract the token kind and cast to
		// the expected type.
		kindv, ok := claims["type"]
		if !ok {
			return nil, jwt.ErrInvalidType
		}
		token.Kind, _ = kindv.(string)

		// extract the token value and cast to
		// expected type.
		textv, ok := claims["text"]
		if !ok {
			return nil, jwt.ErrInvalidType
		}
		token.Text, _ = textv.(string)

		// invoke the callback function to retrieve
		// the secret key used to verify
		secret, err := fn(token)
		return []byte(secret), err
	}
}
