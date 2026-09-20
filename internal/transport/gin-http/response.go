package gin_http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/pkg/logger"
)

const (
	BadRequest = "BAD_REQUEST"
	Internal   = "INTERNAL"
)

type ErrorResponse struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

func NewErrorResponseWithCode(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, ErrorResponse{Code: code, Message: message})
}

func NewErrorResponse(c *gin.Context, statusCode int, message string) {
	switch statusCode {
	case http.StatusBadRequest:
		c.AbortWithStatusJSON(statusCode, ErrorResponse{Code: BadRequest, Message: message})
	default:
		c.AbortWithStatusJSON(statusCode, ErrorResponse{Code: Internal, Message: message})
	}
}

// paramInt64 parses the named path param as int64, logging and responding
// with 400 Bad Request on failure. The bool return reports success so the
// caller can just `if !ok { return }`.
func (h *Handler) paramInt64(c *gin.Context, name, op string) (int64, bool) {
	raw := c.Param(name)
	v, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		logger.From(c.Request.Context()).Warn("invalid path param",
			logger.Op(op), logger.String("param", name), logger.Err(err))
		NewErrorResponse(c, http.StatusBadRequest, "invalid "+name)
		return 0, false
	}
	return v, true
}
