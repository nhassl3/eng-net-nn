package gin_http

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/nhassl3/IpBuild-backend/pkg/logger"
	"github.com/nhassl3/IpBuild-backend/pkg/minio"
)

func (h *Handler) getAllVacancies(c *gin.Context) {
	limit, offset := parseQuery(c)
	vacancies, err := h.services.Vacancies.List(c.Request.Context(), limit, offset)
	if err != nil {
		handleError(c, "getAllVacancies", err)
		return
	}
	c.JSON(http.StatusOK, vacancies)
}

func (h *Handler) getVacancy(c *gin.Context) {
	id := c.Param("id")
	vacancy, err := h.services.Vacancies.GetVacancy(c.Request.Context(), id)
	if err != nil {
		handleError(c, "getVacancy", err)
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
		handleError(c, "createVacancy", err)
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
		handleError(c, "updateVacancy", err)
		return
	}
	c.JSON(http.StatusOK, vacancy)
}

func (h *Handler) deleteVacancy(c *gin.Context) {
	id := c.Param("id")

	if err := h.services.Vacancies.Delete(c.Request.Context(), id); err != nil {
		handleError(c, "deleteVacancy", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) listJd(c *gin.Context) {
	limit, offset := parseQuery(c)
	JDs, err := h.services.Vacancies.ListJd(c.Request.Context(), limit, offset)
	if err != nil {
		handleError(c, "listJd", err)
		return
	}
	c.JSON(http.StatusOK, JDs)
}

func (h *Handler) getJd(c *gin.Context) {
	idInt, ok := h.paramInt32(c, "id", "getJd")
	if !ok {
		return
	}

	jd, err := h.services.Vacancies.GetJd(c.Request.Context(), idInt)
	if err != nil {
		handleError(c, "getJd", err)
		return
	}
	c.JSON(http.StatusOK, jd)
}

func (h *Handler) createJd(c *gin.Context) {
	var input domain.CreateJobDirectionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	jd, err := h.services.Vacancies.CreateJd(c.Request.Context(), &input)
	if err != nil {
		handleError(c, "createJd", err)
		return
	}
	c.JSON(http.StatusCreated, jd)
}

func (h *Handler) updateJd(c *gin.Context) {
	idInt, ok := h.paramInt32(c, "id", "updateJd")
	if !ok {
		return
	}

	var input domain.UpdateJobDirectionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	vacancy, err := h.services.Vacancies.UpdateJd(c.Request.Context(), idInt, &input)
	if err != nil {
		handleError(c, "updateJd", err)
		return
	}
	c.JSON(http.StatusOK, vacancy)
}

func (h *Handler) deleteJd(c *gin.Context) {
	idInt, ok := h.paramInt32(c, "id", "deleteJd")
	if !ok {
		return
	}

	if err := h.services.Vacancies.DeleteJd(c.Request.Context(), idInt); err != nil {
		handleError(c, "deleteJd", err)
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
	if err := json.Unmarshal([]byte(c.PostForm("json")), &input); err != nil {
		logger.From(c.Request.Context()).Warn("respond: decode json field", logger.Err(err))
		NewErrorResponse(c, http.StatusBadRequest, "invalid json body")
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		NewErrorResponse(c, http.StatusBadRequest, "file is required")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		logger.From(c.Request.Context()).Warn("respond: open uploaded file", logger.Err(err))
		NewErrorResponse(c, http.StatusBadRequest, "cannot read uploaded file")
		return
	}
	defer func() {
		_ = file.Close()
	}()

	// Bound the in-memory read so an oversized upload can't exhaust memory; the
	// service rejects anything above the limit via minio.MaxFileSize.
	data, err := io.ReadAll(io.LimitReader(file, minio.MaxFileSize+1))
	if err != nil {
		logger.From(c.Request.Context()).Warn("respond: read uploaded file", logger.Err(err))
		NewErrorResponse(c, http.StatusBadRequest, "cannot read uploaded file")
		return
	}

	dto := domain.FileUploadInput{FileData: data}

	if err := h.services.Vacancies.Respond(
		c.Request.Context(), input.VacancyID, &input.ApplicantsFormInput, &dto,
	); err != nil {
		handleError(c, "respond", err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "response submitted"})
}

func (h *Handler) getRespondVacancies(c *gin.Context) {
	respondVacancies, err := h.services.Vacancies.GetRespondVacancies(c.Request.Context())
	if err != nil {
		handleError(c, "getRespondVacancies", err)
		return
	}
	c.JSON(http.StatusOK, respondVacancies)
}

func (h *Handler) getRespondVacancy(c *gin.Context) {
	id := c.Param("id")
	respondVacancy, err := h.services.Vacancies.GetRespondVacancy(c.Request.Context(), id)
	if err != nil {
		handleError(c, "getRespondVacancy", err)
		return
	}
	c.JSON(http.StatusOK, respondVacancy)
}

func parseQuery(c *gin.Context) (int32, int32) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil {
		limit = 4
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil {
		offset = 0
	}
	return int32(limit), int32(offset)
}
