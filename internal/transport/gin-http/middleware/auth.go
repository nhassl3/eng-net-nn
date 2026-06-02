package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/internal/service"
)

const (
	authorizationHeader = "Authorization"
	userCtx             = "userId"
	roleCtx             = "userRole"
	tokenCtx            = "token"
)

type AuthInterceptor struct {
	s service.Authorization
}

func NewAuthInterceptor(s service.Authorization) *AuthInterceptor {
	return &AuthInterceptor{s: s}
}

func (i *AuthInterceptor) UserIdentity(c *gin.Context) {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		newErrorResponse(c, http.StatusUnauthorized, "no authorization header")
		return
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		newErrorResponse(c, http.StatusUnauthorized, "invalid authorization header format")
		return
	}

	tokenStr := parts[1]
	user, err := i.s.ParseToken(c.Request.Context(), tokenStr)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}
	if user == nil {
		newErrorResponse(c, http.StatusUnauthorized, "invalid token")
		return
	}

	c.Set(userCtx, user.UUID)
	c.Set(roleCtx, user.Role)
	c.Set(tokenCtx, tokenStr)
	c.Next()
}

func (i *AuthInterceptor) AdminIdentity(c *gin.Context) {
	role, exists := c.Get(roleCtx)
	if !exists {
		newErrorResponse(c, http.StatusForbidden, "access denied")
		return
	}

	roleStr, ok := role.(string)
	if !ok || roleStr != "admin" {
		newErrorResponse(c, http.StatusForbidden, "admin access required")
		return
	}

	c.Next()
}

func (i *AuthInterceptor) GetUserIdByToken(c *gin.Context) string {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	tokenStr := parts[1]
	user, err := i.s.ParseToken(c.Request.Context(), tokenStr)
	if err != nil {
		return ""
	}
	if user == nil {
		return ""
	}

	return user.UUID
}

type errorResponse struct {
	Message string `json:"message"`
}

func newErrorResponse(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, errorResponse{message})
}

func GetUserId(c *gin.Context) (string, error) {
	id, ok := c.Get(userCtx)
	if !ok {
		newErrorResponse(c, http.StatusUnauthorized, "user not found in context")
		return "", errors.New("user not found in context")
	}

	idStr, ok := id.(string)
	if !ok || idStr == "" {
		newErrorResponse(c, http.StatusUnauthorized, "user id has invalid type")
		return "", errors.New("user id has invalid type")
	}

	return idStr, nil
}
