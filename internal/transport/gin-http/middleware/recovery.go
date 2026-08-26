package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/pkg/logger"
)

// Recovery replaces gin.Recovery(): it catches panics, logs them at Error
// with a stacktrace and the request's logger (carrying request_id), and
// responds with a plain 500 instead of letting gin print to stderr. It must
// be registered after Logging so the request-scoped logger is already in
// the request context.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.From(c.Request.Context()).Error("panic recovered",
					logger.Any("panic", rec),
					logger.String("method", c.Request.Method),
					logger.String("path", c.Request.URL.Path),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			}
		}()
		c.Next()
	}
}
