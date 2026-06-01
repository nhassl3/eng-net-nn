package gin_http

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
)

func (h *Handler) signUp(c *gin.Context) {
	var input domain.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("signUp: bind json", slog.String("err", err.Error()))
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	user, tokenPair, err := h.services.Authorization.CreateUser(c.Request.Context(), &input)
	if err != nil {
		h.logger.Error("signUp: create user", slog.String("err", err.Error()))
		handleError(c, err)
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
		h.logger.Error("signIn: bind json", slog.String("err", err.Error()))
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.services.Authorization.SignIn(c.Request.Context(), input.Username, input.Password)
	if err != nil {
		h.logger.Error("signIn: sign in", slog.String("err", err.Error()))
		handleError(c, err)
		return
	}

	tokenPair, err := h.services.Authorization.GenerateToken(c.Request.Context(), user)
	if err != nil {
		h.logger.Error("signIn: generate token", slog.String("err", err.Error()))
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":   user,
		"tokens": tokenPair,
	})
}

func (h *Handler) refresh(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" validator:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	tokenPair, err := h.services.Authorization.RefreshToken(c.Request.Context(), input.RefreshToken)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, tokenPair)
}

func (h *Handler) logout(c *gin.Context) {
	header := c.GetHeader("Authorization")
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		NewErrorResponse(c, http.StatusUnauthorized, "invalid authorization header")
		return
	}

	if err := h.services.Authorization.Logout(c.Request.Context(), parts[1]); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
