package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/internal/service"
	"github.com/nhassl3/IpBuild-backend/pkg/logger"
)

const (
	authorizationHeader = "Authorization"
	UserIdCtx           = "userId"
	roleCtx             = "userRole"
	TokenCtx            = "token"
)

type AuthInterceptor struct {
	s service.Authorization
}

func NewAuthInterceptor(s service.Authorization) *AuthInterceptor {
	return &AuthInterceptor{s: s}
}

func (i *AuthInterceptor) UserIdentity(c *gin.Context) {
	log := logger.From(c.Request.Context())

	header := c.GetHeader(authorizationHeader)
	if header == "" {
		log.Warn("auth: missing authorization header")
		newErrorResponse(c, http.StatusUnauthorized, "no authorization header")
		return
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		log.Warn("auth: invalid authorization header format")
		newErrorResponse(c, http.StatusUnauthorized, "invalid authorization header format")
		return
	}

	tokenStr := parts[1]
	user, err := i.s.ParseToken(c.Request.Context(), tokenStr)
	if err != nil {
		log.Warn("auth: parse token failed", logger.Err(err))
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}
	if user == nil {
		log.Warn("auth: token parsed to nil user")
		newErrorResponse(c, http.StatusUnauthorized, "invalid token")
		return
	}

	c.Set(UserIdCtx, user.UUID)
	c.Set(roleCtx, user.Role)
	c.Set(TokenCtx, tokenStr)
	c.Next()
}

func (i *AuthInterceptor) AdminIdentity(c *gin.Context) {
	role, exists := c.Get(roleCtx)
	if !exists {
		logger.From(c.Request.Context()).Warn("auth: admin check without role in context")
		newErrorResponse(c, http.StatusForbidden, "access denied")
		return
	}

	roleStr, ok := role.(string)
	if !ok || roleStr != "admin" {
		logger.From(c.Request.Context()).Warn("auth: admin access denied", logger.String("role", roleStr))
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
