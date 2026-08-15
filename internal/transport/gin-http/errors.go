package gin_http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
	pkgauth "github.com/nhassl3/IpBuild-backend/pkg/auth"
)

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrUserAlreadyExists),
		errors.Is(err, domain.ErrVacanciesAlreadyExists),
		errors.Is(err, domain.ErrVacanciesAlreadyRespond),
		errors.Is(err, domain.ErrPlanRequestAlreadyExists),
		errors.Is(err, domain.ErrRespondAlreadyExists),
		errors.Is(err, domain.ErrVacancyAlreadyExists):
		NewErrorResponse(c, http.StatusConflict, err.Error())

	case errors.Is(err, domain.ErrInvalidCredentials),
		pkgauth.IsAny(err):
		NewErrorResponse(c, http.StatusUnauthorized, err.Error())

	case errors.Is(err, domain.ErrUserNotExists),
		errors.Is(err, domain.ErrVacanciesNotExists),
		errors.Is(err, domain.ErrPlanRequestNotExists),
		errors.Is(err, domain.ErrVacancyNotExists),
		errors.Is(err, domain.ErrDirectionNotFound),
		errors.Is(err, domain.ErrRespondVacanciesNotExists),
		errors.Is(err, domain.ErrRespondVacancyNotExists):
		NewErrorResponse(c, http.StatusNotFound, err.Error())

	case errors.Is(err, domain.ErrFileTooLarge):
		NewErrorResponse(c, http.StatusRequestEntityTooLarge, err.Error())

	case errors.Is(err, domain.ErrInvalidContentType):
		NewErrorResponse(c, http.StatusUnsupportedMediaType, err.Error())

	default:
		NewErrorResponse(c, http.StatusInternalServerError, "internal server error")
	}
}
