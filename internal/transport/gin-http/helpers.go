package gin_http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultLimit = 20
	minLimit     = 1
	maxLimit     = 100
)

// parseQuery parses "limit" and "offset" query params. limit is clamped to
// [minLimit, maxLimit] with a default of defaultLimit, offset must be >= 0.
// Malformed values respond with 400 instead of silently falling back.
func parseQuery(c *gin.Context) (limit, offset int32, ok bool) {
	limitStr := c.DefaultQuery("limit", strconv.Itoa(defaultLimit))
	limitVal, err := strconv.Atoi(limitStr)
	if err != nil || limitVal < minLimit || limitVal > maxLimit {
		NewErrorResponse(c, http.StatusBadRequest, "limit must be an integer between 1 and 100")
		return 0, 0, false
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offsetVal, err := strconv.Atoi(offsetStr)
	if err != nil || offsetVal < 0 {
		NewErrorResponse(c, http.StatusBadRequest, "offset must be a non-negative integer")
		return 0, 0, false
	}

	return int32(limitVal), int32(offsetVal), true
}
