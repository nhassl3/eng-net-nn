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
		errors.Is(err, domain.ErrPlanRequestAlreadyExists):
		NewErrorResponse(c, http.StatusConflict, err.Error())

	case errors.Is(err, domain.ErrInvalidCredentials),
		pkgauth.IsAny(err):
		NewErrorResponse(c, http.StatusUnauthorized, err.Error())

	case errors.Is(err, domain.ErrUserNotExists),
		errors.Is(err, domain.ErrVacanciesNotExists),
		errors.Is(err, domain.ErrPlanRequestNotExists):
		NewErrorResponse(c, http.StatusNotFound, err.Error())

	default:
		NewErrorResponse(c, http.StatusInternalServerError, "internal server error")
	}
}
