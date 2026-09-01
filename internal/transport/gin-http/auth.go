package gin_http

import (
	"errors"
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

	h.setRefreshCookie(c, tokenPair.RefreshToken)

	c.JSON(http.StatusCreated, gin.H{
		"user":         user,
		"access_token": tokenPair.AccessToken,
		"expires_in":   tokenPair.ExpiresIn,
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

	user, tokenPair, err := h.services.Authorization.SignIn(c.Request.Context(), &input)
	if err != nil {
		handleError(c, "signIn", err)
		return
	}

	h.setRefreshCookie(c, tokenPair.RefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"user":         user,
		"access_token": tokenPair.AccessToken,
		"expires_in":   tokenPair.ExpiresIn,
	})
}

func (h *Handler) refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(h.tokenCfg.Cookie.Name)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			var input domain.RefreshInput
			if err := c.ShouldBindJSON(&input); err != nil {
				handleError(c, "refresh: BODY", domain.ErrInvalidToken)
				return
			}
			refreshToken = input.RefreshToken
		} else {
			handleError(c, "refresh", err)
			return
		}
	}

	tokenPair, err := h.services.Authorization.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		handleError(c, "refresh", err)
		return
	}

	h.setRefreshCookie(c, tokenPair.RefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"access_token": tokenPair.AccessToken,
		"expires_in":   tokenPair.ExpiresIn,
	})
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
	refreshToken, _ := c.Cookie(h.tokenCfg.Cookie.Name)
	if err := h.services.Authorization.Logout(c.Request.Context(), c.GetString(middleware.TokenCtx), refreshToken); err != nil {
		handleError(c, "logout", err)
		return
	}

	h.clearRefreshCookie(c)

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *Handler) setRefreshCookie(c *gin.Context, token string) {
	setSameSite(c, h.tokenCfg.Cookie.SameSite)

	c.SetCookie(
		h.tokenCfg.Cookie.Name,
		token,
		int(h.tokenCfg.RefreshTTL.Seconds()),
		h.tokenCfg.Cookie.Path,
		h.tokenCfg.Cookie.Domain,
		h.tokenCfg.Cookie.Secure,
		true,
	)
}

func (h *Handler) clearRefreshCookie(c *gin.Context) {
	setSameSite(c, h.tokenCfg.Cookie.SameSite)
	c.SetCookie(
		h.tokenCfg.Cookie.Name,
		"",
		-1, // real remove token from cookie immediately
		h.tokenCfg.Cookie.Path,
		h.tokenCfg.Cookie.Domain,
		h.tokenCfg.Cookie.Secure,
		true,
	)
}

func setSameSite(c *gin.Context, sameSite string) {
	switch sameSite {
	case "lax":
		c.SetSameSite(http.SameSiteLaxMode)
	default:
		c.SetSameSite(http.SameSiteNoneMode)
	}
}
