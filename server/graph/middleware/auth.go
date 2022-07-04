package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/shared/token"
)

// A private key for context that only this package can access. This is important
// to prevent collisions between different context uses
const userCtxKey = "user"

func SetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user *model.User

		_, err := token.ParseRequest(c.Request, func(t *token.Token) (string, error) {
			var err error
			user, err = store.FromContext(c).GetUserLogin(t.Text)
			return user.Hash, err
		})
		if err == nil {
			confv := c.MustGet("config")
			if conf, ok := confv.(*model.Settings); ok {
				user.Admin = conf.IsAdmin(user)
			}
			ctx := context.WithValue(c.Request.Context(), userCtxKey, user)
			c.Request = c.Request.WithContext(ctx)

			// TODO
			// // if this is a session token (ie not the API token)
			// // this means the user is accessing with a web browser,
			// // so we should implement CSRF protection measures.
			// if t.Kind == token.SessToken {
			// 	err = token.CheckCsrf(c.Request, func(t *token.Token) (string, error) {
			// 		return user.Hash, nil
			// 	})
			// 	// if csrf token validation fails, exit immediately
			// 	// with a not authorized error.
			// 	if err != nil {
			// 		c.AbortWithStatus(http.StatusUnauthorized)
			// 		return
			// 	}
			// }
		}
		c.Next()
	}
}

func User(ctx context.Context) *model.User {
	raw, _ := ctx.Value(userCtxKey).(*model.User)
	return raw
}
