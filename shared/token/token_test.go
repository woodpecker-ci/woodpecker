package token_test

import (
	"testing"

	"github.com/franela/goblin"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/shared/token"
)

func TestToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	g := goblin.Goblin(t)
	g.Describe("Token", func() {
		jwtSecret := "secret-to-sign-the-token"

		g.It("should parse a valid token", func() {
			_token := token.New(token.UserToken)
			_token.Set("user-id", "1")
			signedToken, err := _token.Sign(jwtSecret)
			assert.NoError(g, err)

			parsed, err := token.Parse([]token.Type{token.UserToken}, signedToken, func(_ *token.Token) (string, error) {
				return jwtSecret, nil
			})

			assert.NoError(g, err)
			assert.NotNil(g, parsed)
			assert.Equal(g, "1", parsed.Get("user-id"))
		})

		g.It("should fail to parse a token with a wrong type", func() {
			_token := token.New(token.UserToken)
			_token.Set("user-id", "1")
			signedToken, err := _token.Sign(jwtSecret)
			assert.NoError(g, err)

			_, err = token.Parse([]token.Type{token.AgentToken}, signedToken, func(_ *token.Token) (string, error) {
				return jwtSecret, nil
			})

			assert.ErrorIs(g, err, jwt.ErrInvalidType)
		})

		g.It("should fail to parse a token with a wrong secret", func() {
			_token := token.New(token.UserToken)
			_token.Set("user-id", "1")
			signedToken, err := _token.Sign(jwtSecret)
			assert.NoError(g, err)

			_, err = token.Parse([]token.Type{token.UserToken}, signedToken, func(_ *token.Token) (string, error) {
				return "this-is-a-wrong-secret", nil
			})

			assert.ErrorIs(g, err, jwt.ErrSignatureInvalid)
		})
	})
}
