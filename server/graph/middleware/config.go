package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

// TODO: remove
// map config from gin context to "normal" context
func SetConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		config := c.MustGet("config")
		ctx := context.WithValue(c.Request.Context(), "config", config)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
