package gin_http

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/internal/service"
	"github.com/nhassl3/IpBuild-backend/internal/transport/gin-http/interceptors"
	sloggin "github.com/samber/slog-gin"
)

type Handler struct {
	services   *service.Service
	logger     *slog.Logger
	middleware *interceptors.AuthInterceptor
}

func NewHandler(services *service.Service, logger *slog.Logger) *Handler {
	return &Handler{
		services:   services,
		logger:     logger,
		middleware: interceptors.NewAuthInterceptor(services.Authorization),
	}
}

func (h *Handler) InitRoutes(env string) *gin.Engine {
	if env != "local" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(sloggin.New(h.logger))

	auth := router.Group("/auth")
	{
		auth.POST("/signup", h.signUp)
		auth.POST("/login", h.signIn)
		auth.POST("/refresh", h.refresh)
	}

	api := router.Group("/api", h.middleware.UserIdentity)
	{
		api.POST("/logout", h.logout)

		vacancies := api.Group("/vacancies")
		{
			vacancies.GET("/", h.getAllVacancies)
			vacancies.GET("/:id", h.getVacancy)
			vacancies.POST("/", h.middleware.AdminIdentity, h.createVacancy)
			vacancies.PUT("/:id", h.middleware.AdminIdentity, h.updateVacancy)
			vacancies.DELETE("/:id", h.middleware.AdminIdentity, h.deleteVacancy)
			vacancies.POST("/respond", h.respond)
		}

		plan := api.Group("/plan")
		{
			plan.POST("/", h.requestPlan)
			plan.GET("/:id", h.getResponseFromRequest)
		}
	}

	return router
}
