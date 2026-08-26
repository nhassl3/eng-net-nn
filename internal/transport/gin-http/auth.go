package gin_http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/nhassl3/IpBuild-backend/internal/transport/gin-http/middleware"
	"github.com/nhassl3/IpBuild-backend/pkg/logger"
)

func (h *Handler) signUp(c *gin.Context) {
	var input domain.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.From(c.Request.Context()).Warn("signUp: bind json", logger.Err(err))
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	user, tokenPair, err := h.services.Authorization.CreateUser(c.Request.Context(), &input)
	if err != nil {
		handleError(c, "signUp", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":   user,
		"tokens": tokenPair,
	})
}

func (h *Handler) signIn(c *gin.Context) {
	var input domain.SignInInput
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.From(c.Request.Context()).Warn("signIn: bind json", logger.Err(err))
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if input.Username == "" && input.Email == "" && input.ID == "" {
		NewErrorResponse(c, http.StatusBadRequest, "username or email are required")
		return
	}

	user, err := h.services.Authorization.SignIn(c.Request.Context(), &input)
	if err != nil {
		handleError(c, "signIn", err)
		return
	}

	tokenPair, err := h.services.Authorization.GenerateToken(c.Request.Context(), user)
	if err != nil {
		handleError(c, "signIn.GenerateToken", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":   user,
		"tokens": tokenPair,
	})
}

func (h *Handler) refresh(c *gin.Context) {
	var input domain.RefreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	tokenPair, err := h.services.Authorization.RefreshToken(c.Request.Context(), input.RefreshToken)
	if err != nil {
		handleError(c, "refresh", err)
		return
	}

	c.JSON(http.StatusOK, tokenPair)
}

func (h *Handler) me(c *gin.Context) {
	user, err := h.services.Authorization.GetMe(c.Request.Context(), c.GetString(middleware.UserIdCtx))
	if err != nil {
		handleError(c, "me", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *Handler) logout(c *gin.Context) {
	if err := h.services.Authorization.Logout(c.Request.Context(), c.GetString(middleware.TokenCtx)); err != nil {
		handleError(c, "logout", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
