package gin_http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/nhassl3/IpBuild-backend/pkg/logger/sl"
)

func (h *Handler) requestPlan(c *gin.Context) {
	var input domain.CreatePlanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var userId *string
	if id := h.middleware.GetUserIdByToken(c); id != "" {
		userId = &id
	}

	plan, err := h.services.Plan.CreatePlan(c.Request.Context(), &input, userId)
	if err != nil {
		h.logger.Error("requestPlan", sl.ErrLog(err))
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, plan)
}

func (h *Handler) getResponseFromRequest(c *gin.Context) {
	id := c.Param("id")

	userPlan, err := h.services.Plan.GetPlan(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("getResponseFromRequest", sl.ErrLog(err))
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, userPlan)
}

func (h *Handler) getAllPlans(c *gin.Context) {
	allPlans, err := h.services.Plan.GetAllPlans(c.Request.Context())
	if err != nil {
		h.logger.Error("getAllPlans", sl.ErrLog(err))
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, allPlans)
}

func (h *Handler) getPlan(c *gin.Context) {
	id := c.Param("id")
	userPlan, err := h.services.Plan.GetPlan(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("getPlan", sl.ErrLog(err))
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, userPlan)
}
