package gin_http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
)

func TestHandleError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"user already exists", domain.ErrUserAlreadyExists, http.StatusConflict, domain.UserAlreadyExists},
		{"vacancies already exist", domain.ErrVacanciesAlreadyExists, http.StatusConflict, domain.VacanciesAlreadyExist},
		{"vacancies already respond", domain.ErrVacanciesAlreadyRespond, http.StatusConflict, domain.VacanciesAlreadyRespond},
		{"plan request already exists", domain.ErrPlanRequestAlreadyExists, http.StatusConflict, domain.PlanRequestAlreadyExists},
		{"respond already exists", domain.ErrRespondAlreadyExists, http.StatusConflict, domain.RespondAlreadyExists},
		{"vacancy already exists", domain.ErrVacancyAlreadyExists, http.StatusConflict, domain.VacancyAlreadyExists},
		{"direction has vacancies", domain.ErrDirectionHasVacancies, http.StatusConflict, domain.DirectionHasVacancies},

		{"invalid credentials", domain.ErrInvalidCredentials, http.StatusUnauthorized, domain.InvalidCredentials},
		{"invalid token", domain.ErrInvalidToken, http.StatusUnauthorized, domain.InvalidToken},
		{"token revoked", domain.ErrTokenRevoked, http.StatusUnauthorized, domain.TokenRevoked},
		{"expired token", domain.ErrExpiredToken, http.StatusUnauthorized, domain.TokenExpired},
		{"unauthorized", domain.ErrUnauthorized, http.StatusUnauthorized, domain.Unauthorized},

		{"user not exists", domain.ErrUserNotExists, http.StatusNotFound, domain.UserNotFound},
		{"vacancies not exists", domain.ErrVacanciesNotExists, http.StatusNotFound, domain.VacanciesNotFound},
		{"plan request not exists", domain.ErrPlanRequestNotExists, http.StatusNotFound, domain.PlanRequestNotFound},
		{"vacancy not exists", domain.ErrVacancyNotExists, http.StatusNotFound, domain.VacancyNotFound},
		{"direction not found", domain.ErrDirectionNotFound, http.StatusNotFound, domain.DirectionNotFound},
		{"respond vacancies not exists", domain.ErrRespondVacanciesNotExists, http.StatusNotFound, domain.RespondVacanciesNotFound},
		{"respond vacancy not exists", domain.ErrRespondVacancyNotExists, http.StatusNotFound, domain.RespondVacancyNotFound},

		{"file too large", domain.ErrFileTooLarge, http.StatusRequestEntityTooLarge, domain.FileTooLarge},
		{"invalid content type", domain.ErrInvalidContentType, http.StatusUnsupportedMediaType, domain.InvalidContentType},

		{"invalid param", domain.ErrInvalidParam, http.StatusBadRequest, domain.InvalidParam},
		{"empty data", domain.ErrEmptyData, http.StatusBadRequest, domain.EmptyData},
		{"invalid input", domain.ErrInvalidInput, http.StatusBadRequest, domain.InvalidInput},

		{"unknown non-domain error", errors.New("boom"), http.StatusInternalServerError, Internal},
		{"unmapped domain error falls into default", domain.ErrRedisNotFound, http.StatusInternalServerError, Internal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			handleError(c, "test.op", tt.err)

			if w.Code != tt.wantStatus {
				t.Fatalf("status: expected %d, got %d, body: %s", tt.wantStatus, w.Code, w.Body.String())
			}

			var resp ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to decode response body: %v, body: %s", err, w.Body.String())
			}

			if resp.Code != tt.wantCode {
				t.Fatalf("code: expected %q, got %q", tt.wantCode, resp.Code)
			}
		})
	}
}
