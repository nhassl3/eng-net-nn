package gin_http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
