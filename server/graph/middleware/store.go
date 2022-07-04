package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// TODO: remove
// map store from gin context to "normal" context
func SetStore() gin.HandlerFunc {
	return func(c *gin.Context) {
		store := store.FromContext(c)
		ctx := context.WithValue(c.Request.Context(), "store", store)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
