package gin_http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/internal/config"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/nhassl3/IpBuild-backend/internal/service"
	"github.com/nhassl3/IpBuild-backend/pkg/logger"
)

// mockAuthorization is a no-op stand-in for service.Authorization used to
// verify that handlers never reach the service layer when input binding
// fails — every method fails the test if called.
type mockAuthorization struct {
	t *testing.T
}

func (m *mockAuthorization) CreateUser(context.Context, *domain.CreateUserInput) (*domain.User, *domain.TokenPair, error) {
	m.t.Fatal("CreateUser must not be called when JSON binding fails")
	return nil, nil, nil
}

func (m *mockAuthorization) SignIn(context.Context, *domain.SignInInput) (*domain.User, *domain.TokenPair, error) {
	m.t.Fatal("SignIn must not be called when JSON binding fails")
	return nil, nil, nil
}

func (m *mockAuthorization) GenerateToken(context.Context, *domain.User) (*domain.TokenPair, error) {
	m.t.Fatal("GenerateToken must not be called")
	return nil, nil
}

func (m *mockAuthorization) ParseToken(context.Context, string) (*domain.User, error) {
	m.t.Fatal("ParseToken must not be called")
	return nil, nil
}

func (m *mockAuthorization) RefreshToken(context.Context, string) (*domain.TokenPair, error) {
	m.t.Fatal("RefreshToken must not be called")
	return nil, nil
}

func (m *mockAuthorization) Logout(context.Context, string, string) error {
	m.t.Fatal("Logout must not be called")
	return nil
}

func (m *mockAuthorization) GetMe(context.Context, string) (*domain.User, error) {
	m.t.Fatal("GetMe must not be called")
	return nil, nil
}

func newTestHandler(t *testing.T) *Handler {
	gin.SetMode(gin.TestMode)

	log, err := logger.New(logger.Config{})
	if err != nil {
		t.Fatalf("failed to build test logger: %v", err)
	}

	services := &service.Service{Authorization: &mockAuthorization{t: t}}

	return NewHandler(services, log, &config.Token{})
}

// TestSignUp_NoData covers issue #19: an empty JSON body must be rejected by
// binding validation with 400, before the request ever reaches the service.
func TestSignUp_NoData(t *testing.T) {
	h := newTestHandler(t)

	router := gin.New()
	router.POST("/auth/signup", h.signUp)

	req := httptest.NewRequest(http.MethodPost, "/auth/signup", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}
