package token_test

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/shared/token"
)

const jwtSecret = "secret-to-sign-the-token"

func TestTokenValid(t *testing.T) {
	_token := token.New(token.UserToken)
	_token.Set("user-id", "1")
	signedToken, err := _token.Sign(jwtSecret)
	assert.NoError(t, err)

	parsed, err := token.Parse([]token.Type{token.UserToken}, signedToken, func(_ *token.Token) (string, error) {
		return jwtSecret, nil
	})

	assert.NoError(t, err)
	assert.NotNil(t, parsed)
	assert.Equal(t, "1", parsed.Get("user-id"))
}

func TestTokenWrongType(t *testing.T) {
	_token := token.New(token.UserToken)
	_token.Set("user-id", "1")
	signedToken, err := _token.Sign(jwtSecret)
	assert.NoError(t, err)

	_, err = token.Parse([]token.Type{token.AgentToken}, signedToken, func(_ *token.Token) (string, error) {
		return jwtSecret, nil
	})

	assert.ErrorIs(t, err, jwt.ErrInvalidType)
}

func TestTokenWrongSecret(t *testing.T) {
	_token := token.New(token.UserToken)
	_token.Set("user-id", "1")
	signedToken, err := _token.Sign(jwtSecret)
	assert.NoError(t, err)

	_, err = token.Parse([]token.Type{token.UserToken}, signedToken, func(_ *token.Token) (string, error) {
		return "this-is-a-wrong-secret", nil
	})

	assert.ErrorIs(t, err, jwt.ErrSignatureInvalid)
}
