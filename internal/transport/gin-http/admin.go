package gin_http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type searchUserQuery struct {
	Username string `form:"username" binding:"required,min=3,max=30"`
}

func (h *Handler) searchUser(c *gin.Context) {
	var query searchUserQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.services.Admin.SearchUserByUsername(c.Request.Context(), query.Username)
	if err != nil {
		handleError(c, "searchUser", fmt.Errorf("handler.admin: searchUser: failed to find user: %w", err))
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) addAdmin(c *gin.Context) {
	uid := c.Param("id")
	if err := h.services.AddAdmin(c.Request.Context(), uid); err != nil {
		handleError(c, "addAdmin", fmt.Errorf("handler.admin: addAdmin: failed to create an admin: %w", err))
		return
	}
	c.Status(http.StatusCreated)
}

func (h *Handler) addPartner(c *gin.Context) {
	uid := c.Param("id")
	if err := h.services.AddPartner(c.Request.Context(), uid); err != nil {
		handleError(c, "addPartner", fmt.Errorf("handler.partner: addPartner: failed to create an partner: %w", err))
		return
	}
	c.Status(http.StatusCreated)
}
