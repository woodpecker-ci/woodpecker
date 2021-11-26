package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Logger returns a gin.HandlerFunc (middleware) that logs requests using zerolog.
//
// Requests with errors are logged using log.Err().
// Requests without errors are logged using log.Info().
//
// It receives:
//   1. A time package format string (e.g. time.RFC3339).
//   2. A boolean stating whether to use UTC time zone or local.
func Logger(timeFormat string, utc bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		if utc {
			end = end.UTC()
		}

		entry := map[string]interface{}{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
			"latency":    latency,
			"user-agent": c.Request.UserAgent(),
			"time":       end.Format(timeFormat),
		}

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			log.Error().Str("error", c.Errors.String()).Fields(entry).Msg("")
		} else {
			log.Info().Fields(entry).Msg("")
		}
	}
}
