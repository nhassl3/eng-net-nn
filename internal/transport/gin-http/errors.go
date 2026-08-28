package gin_http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
	pkgauth "github.com/nhassl3/IpBuild-backend/pkg/auth"
	"github.com/nhassl3/IpBuild-backend/pkg/logger"
)

// handleError maps a service-layer error to an HTTP response and logs it
// exactly once, using the request-scoped logger from ctx. Expected domain
// errors (conflict/not-found/etc.) log at Warn; anything unrecognized is an
// internal error and logs at Error. op identifies the handler operation,
// e.g. "vacancy.Create".
func handleError(c *gin.Context, op string, err error) {
	log := logger.From(c.Request.Context())

	switch {
	case errors.Is(err, domain.ErrUserAlreadyExists),
		errors.Is(err, domain.ErrVacanciesAlreadyExists),
		errors.Is(err, domain.ErrVacanciesAlreadyRespond),
		errors.Is(err, domain.ErrPlanRequestAlreadyExists),
		errors.Is(err, domain.ErrRespondAlreadyExists),
		errors.Is(err, domain.ErrVacancyAlreadyExists),
		errors.Is(err, domain.ErrDirectionHasVacancies):
		log.Warn("request rejected: conflict", logger.Op(op), logger.Err(err))
		NewErrorResponse(c, http.StatusConflict, errString(err.Error()))

	case errors.Is(err, domain.ErrInvalidCredentials),
		pkgauth.IsAny(err):
		log.Warn("request rejected: unauthorized", logger.Op(op), logger.Err(err))
		NewErrorResponse(c, http.StatusUnauthorized, errString(err.Error()))

	case errors.Is(err, domain.ErrUserNotExists),
		errors.Is(err, domain.ErrVacanciesNotExists),
		errors.Is(err, domain.ErrPlanRequestNotExists),
		errors.Is(err, domain.ErrVacancyNotExists),
		errors.Is(err, domain.ErrDirectionNotFound),
		errors.Is(err, domain.ErrRespondVacanciesNotExists),
		errors.Is(err, domain.ErrRespondVacancyNotExists):
		log.Warn("request rejected: not found", logger.Op(op), logger.Err(err))
		NewErrorResponse(c, http.StatusNotFound, errString(err.Error()))

	case errors.Is(err, domain.ErrFileTooLarge):
		log.Warn("request rejected: file too large", logger.Op(op), logger.Err(err))
		NewErrorResponse(c, http.StatusRequestEntityTooLarge, errString(err.Error()))

	case errors.Is(err, domain.ErrInvalidContentType):
		log.Warn("request rejected: invalid content type", logger.Op(op), logger.Err(err))
		NewErrorResponse(c, http.StatusUnsupportedMediaType, errString(err.Error()))

	default:
		log.Error("request failed: unhandled error", logger.Op(op), logger.Err(err))
		NewErrorResponse(c, http.StatusInternalServerError, "internal server error")
	}
}

func errString(errStr string) string {
	return errStr[1+strings.LastIndex(errStr, ":"):]
}
