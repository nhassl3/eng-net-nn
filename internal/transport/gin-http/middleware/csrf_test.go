package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

var allowed = []string{"https://app.example", "http://localhost:5173"}

func run(t *testing.T, handlers []gin.HandlerFunc, headers map[string]string) int {
	t.Helper()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/refresh", append(handlers, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})...)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w.Code
}

func TestCrossSiteGuard(t *testing.T) {
	guard := []gin.HandlerFunc{CrossSiteGuard(allowed)}

	tests := []struct {
		name    string
		headers map[string]string
		want    int
	}{
		{"allowed origin", map[string]string{"Origin": "https://app.example"}, http.StatusOK},
		{"allowed origin trailing slash", map[string]string{"Origin": "https://app.example/"}, http.StatusOK},
		{"foreign origin", map[string]string{"Origin": "https://evil.example"}, http.StatusForbidden},
		{"scheme mismatch", map[string]string{"Origin": "http://app.example"}, http.StatusForbidden},
		{"cross-site form post without origin", map[string]string{"Sec-Fetch-Site": "cross-site"}, http.StatusForbidden},
		{"same-origin fetch metadata", map[string]string{"Sec-Fetch-Site": "same-origin"}, http.StatusOK},
		{"non-browser client", nil, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := run(t, guard, tt.headers); got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRequireRequestedWith(t *testing.T) {
	chain := []gin.HandlerFunc{RequireRequestedWith}

	tests := []struct {
		name    string
		headers map[string]string
		want    int
	}{
		{"header present", map[string]string{"X-Requested-With": RequestedWithValue}, http.StatusOK},
		{"header missing", nil, http.StatusForbidden},
		{"wrong value", map[string]string{"X-Requested-With": "XMLHttpRequest"}, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := run(t, chain, tt.headers); got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
