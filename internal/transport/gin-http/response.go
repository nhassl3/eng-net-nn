package gin_http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/pkg/logger/sl"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func NewErrorResponse(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, ErrorResponse{message})
}

// paramInt32 parses the named path param as int32, logging and responding
// with 400 Bad Request on failure. The bool return reports success so the
// caller can just `if !ok { return }`.
func (h *Handler) paramInt32(c *gin.Context, name, op string) (int32, bool) {
	raw := c.Param(name)
	v, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		h.logger.Error(op+": invalid path param", slog.String("param", name), sl.ErrLog(err))
		NewErrorResponse(c, http.StatusBadRequest, "invalid "+name)
		return 0, false
	}
	return int32(v), true
}
