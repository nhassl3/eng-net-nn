package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nhassl3/IpBuild-backend/pkg/logger"
)

// RequestIDHeader is the header used to propagate/accept a request ID.
const RequestIDHeader = "X-Request-ID"

// requestIDCtx is the gin context key the request ID is stored under.
const requestIDCtx = "requestId"

// RequestID reads the inbound X-Request-ID header, or generates a new one,
// stores it on the gin context and echoes it back on the response so callers
// (and downstream logs) can correlate a request end-to-end.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(requestIDCtx, id)
		c.Header(RequestIDHeader, id)
		c.Next()
	}
}

// Logging injects a request-scoped logger (carrying request_id) into the
// request context, then emits a single structured access-log line once the
// handler chain completes. It replaces gin's own logger entirely — gin never
// writes to stdout on its own.
func Logging(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID, _ := c.Get(requestIDCtx)
		requestIDStr, _ := requestID.(string)

		reqLog := log.Named("http").With(logger.RequestID(requestIDStr))
		c.Request = c.Request.WithContext(logger.Inject(c.Request.Context(), reqLog))

		c.Next()

		status := c.Writer.Status()
		fields := []logger.Field{
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.String("route", c.FullPath()),
			logger.Int("status", status),
			logger.Duration(time.Since(start)),
			logger.Int("bytes", c.Writer.Size()),
			logger.String("client_ip", c.ClientIP()),
			logger.String("user_agent", c.Request.UserAgent()),
		}
		if userID := c.GetString(UserIdCtx); userID != "" {
			fields = append(fields, logger.UserID(userID))
		}
		if len(c.Errors) > 0 {
			fields = append(fields, logger.String("errors", c.Errors.String()))
		}

		switch {
		case status >= 500:
			reqLog.Error("http request", fields...)
		case status >= 400:
			reqLog.Warn("http request", fields...)
		default:
			reqLog.Info("http request", fields...)
		}
	}
}
