package gin_http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
)

func (h *Handler) requestPlan(c *gin.Context) {
	var input domain.CreatePlanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	plan, err := h.services.Plan.CreatePlan(c.Request.Context(), &input)
	if err != nil {
		h.logger.Error("requestPlan", slog.String("err", err.Error()))
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, plan)
}

func (h *Handler) getResponseFromRequest(c *gin.Context) {
	id := c.Param("id")

	userPlan, err := h.services.Plan.GetPlan(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("getResponseFromRequest", slog.String("err", err.Error()))
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, userPlan)
}
