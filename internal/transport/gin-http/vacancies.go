package gin_http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/nhassl3/IpBuild-backend/pkg/logger/sl"
)

func (h *Handler) getAllVacancies(c *gin.Context) {
	vacancies, err := h.services.Vacancies.List(c.Request.Context())
	if err != nil {
		h.logger.Error("getAllVacancies", slog.String("err", err.Error()))
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, vacancies)
}

func (h *Handler) getVacancy(c *gin.Context) {
	id := c.Param("id")
	vacancy, err := h.services.Vacancies.GetVacancy(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("getVacancy", slog.String("err", err.Error()))
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, vacancy)
}

func (h *Handler) createVacancy(c *gin.Context) {
	var input domain.CreateVacancyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	vacancy, err := h.services.Vacancies.Create(c.Request.Context(), &input)
	if err != nil {
		h.logger.Error("createVacancy", slog.String("err", err.Error()))
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, vacancy)
}

func (h *Handler) updateVacancy(c *gin.Context) {
	id := c.Param("id")

	var input domain.UpdatedVacancyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	vacancy, err := h.services.Vacancies.Update(c.Request.Context(), id, &input)
	if err != nil {
		h.logger.Error("updateVacancy", slog.String("err", err.Error()))
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, vacancy)
}

func (h *Handler) deleteVacancy(c *gin.Context) {
	id := c.Param("id")

	if err := h.services.Vacancies.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("deleteVacancy", slog.String("err", err.Error()))
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// respond a handler which creates email for owner with next
func (h *Handler) respond(c *gin.Context) {
	var input struct {
		VacancyID string `json:"vacancy_id" validator:"required"`
		domain.ApplicantsFormInput
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.services.Vacancies.Respond(c.Request.Context(), input.VacancyID, &input.ApplicantsFormInput); err != nil {
		h.logger.Error("respond", slog.String("err", err.Error()))
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "response submitted"})
}

func (h *Handler) getRespondVacancies(c *gin.Context) {
	respondVacancies, err := h.services.Vacancies.GetRespondVacancies(c.Request.Context())
	if err != nil {
		h.logger.Error("getRespondVacancies", sl.ErrLog(err))
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, respondVacancies)
}

func (h *Handler) getRespondVacancy(c *gin.Context) {
	id := c.Param("id")
	respondVacancy, err := h.services.Vacancies.GetRespondVacancy(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("getRespondVacancy", sl.ErrLog(err))
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, respondVacancy)
}
