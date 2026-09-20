package gin_http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/nhassl3/IpBuild-backend/internal/transport/gin-http/middleware"
)

func (h *Handler) requestPlan(c *gin.Context) {
	var input domain.CreatePlanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var userUID *string
	if userUid := c.GetString(middleware.UserIdCtx); userUid != "" {
		userUID = &userUid
	}

	plan, err := h.services.Plan.CreatePlan(c.Request.Context(), &input, userUID)
	if err != nil {
		handleError(c, "requestPlan", err)
		return
	}

	c.JSON(http.StatusCreated, plan)
}

func (h *Handler) getResponseFromRequest(c *gin.Context) {
	planUID := c.Param("id")

	userUID := c.GetString(middleware.UserIdCtx)
	if userUID == "" {
		handleError(c, "getResponseFromRequest", domain.ErrUnauthorized)
		return
	}

	userPlan, err := h.services.Plan.GetUserPlan(c.Request.Context(), planUID, userUID)
	if err != nil {
		handleError(c, "getResponseFromRequest", err)
		return
	}
	c.JSON(http.StatusOK, userPlan)
}

func (h *Handler) getAllPlans(c *gin.Context) {
	limit, offset, ok := parseQuery(c)
	if !ok {
		return
	}
	allPlans, err := h.services.Plan.GetAllPlans(c.Request.Context(), limit, offset)
	if err != nil {
		handleError(c, "getAllPlans", err)
		return
	}
	c.JSON(http.StatusOK, allPlans)
}

func (h *Handler) getPlan(c *gin.Context) {
	id := c.Param("id")
	userPlan, err := h.services.Plan.GetPlan(c.Request.Context(), id)
	if err != nil {
		handleError(c, "getPlan", err)
		return
	}
	c.JSON(http.StatusOK, userPlan)
}

func (h *Handler) responseToPlan(c *gin.Context) {
	var input domain.ResponseToPlanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	updatedPlan, err := h.services.Plan.ResponseToPlan(c.Request.Context(), input.PlanUID, input.Message)
	if err != nil {
		handleError(c, "responseToPlan", err)
		return
	}
	c.JSON(http.StatusOK, updatedPlan)
}
