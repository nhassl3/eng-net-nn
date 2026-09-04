package gin_http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/nhassl3/IpBuild-backend/pkg/logger"
)

// handleError maps a service-layer error to an HTTP response and logs it
// exactly once, using the request-scoped logger from ctx. Expected domain
// errors (conflict/not-found/etc.) log at Warn; anything unrecognized is an
// internal error and logs at Error. op identifies the handler operation,
// e.g. "vacancy.Create".
func handleError(c *gin.Context, op string, err error) {
	log := logger.From(c.Request.Context())

	dmnErr, ok := errors.AsType[*domain.DomainError](err)
	if !ok {
		log.Error("request failed: unhandled error", logger.Op(op), logger.Err(err))
		NewErrorResponseWithCode(c, http.StatusInternalServerError, Internal, "internal server error")
		return
	}

	switch {
	case errors.Is(dmnErr, domain.ErrUserAlreadyExists),
		errors.Is(dmnErr, domain.ErrVacanciesAlreadyExists),
		errors.Is(dmnErr, domain.ErrVacanciesAlreadyRespond),
		errors.Is(dmnErr, domain.ErrPlanRequestAlreadyExists),
		errors.Is(dmnErr, domain.ErrRespondAlreadyExists),
		errors.Is(dmnErr, domain.ErrVacancyAlreadyExists),
		errors.Is(dmnErr, domain.ErrDirectionHasVacancies):
		log.Warn("request rejected: conflict", logger.Op(op), logger.Err(dmnErr))
		NewErrorResponseWithCode(c, http.StatusConflict, dmnErr.Code(), errString(dmnErr.Error()))

	case errors.Is(dmnErr, domain.ErrInvalidCredentials),
		errors.Is(dmnErr, domain.ErrInvalidToken),
		errors.Is(dmnErr, domain.ErrTokenRevoked),
		errors.Is(dmnErr, domain.ErrExpiredToken):
		log.Warn("request rejected: unauthorized", logger.Op(op), logger.Err(dmnErr))
		NewErrorResponseWithCode(c, http.StatusUnauthorized, dmnErr.Code(), errString(dmnErr.Error()))

	case errors.Is(dmnErr, domain.ErrUserNotExists),
		errors.Is(dmnErr, domain.ErrVacanciesNotExists),
		errors.Is(dmnErr, domain.ErrPlanRequestNotExists),
		errors.Is(dmnErr, domain.ErrVacancyNotExists),
		errors.Is(dmnErr, domain.ErrDirectionNotFound),
		errors.Is(dmnErr, domain.ErrRespondVacanciesNotExists),
		errors.Is(dmnErr, domain.ErrRespondVacancyNotExists):
		log.Warn("request rejected: not found", logger.Op(op), logger.Err(dmnErr))
		NewErrorResponseWithCode(c, http.StatusNotFound, dmnErr.Error(), errString(dmnErr.Error()))

	case errors.Is(dmnErr, domain.ErrFileTooLarge):
		log.Warn("request rejected: file too large", logger.Op(op), logger.Err(dmnErr))
		NewErrorResponseWithCode(c, http.StatusRequestEntityTooLarge, dmnErr.Code(), errString(dmnErr.Error()))

	case errors.Is(dmnErr, domain.ErrInvalidContentType):
		log.Warn("request rejected: invalid content type", logger.Op(op), logger.Err(dmnErr))
		NewErrorResponseWithCode(c, http.StatusUnsupportedMediaType, dmnErr.Code(), errString(dmnErr.Error()))

	case errors.Is(dmnErr, domain.ErrInvalidParam), errors.Is(dmnErr, domain.ErrEmptyData):
		log.Warn("request rejected: invalid parameter", logger.Op(op), logger.Err(dmnErr))
		NewErrorResponseWithCode(c, http.StatusBadRequest, dmnErr.Code(), errString(dmnErr.Error()))

	default:
		log.Error("request failed: unhandled error", logger.Op(op), logger.Err(dmnErr))
		NewErrorResponseWithCode(c, http.StatusInternalServerError, Internal, "internal server error")
	}
}

func errString(errStr string) string {
	return errStr[1+strings.LastIndex(errStr, ":"):]
}
