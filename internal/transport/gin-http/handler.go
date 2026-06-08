package gin_http

import (
	"log/slog"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/nhassl3/IpBuild-backend/internal/service"
	"github.com/nhassl3/IpBuild-backend/internal/transport/gin-http/middleware"
	valid "github.com/nhassl3/IpBuild-backend/internal/transport/gin-http/validator"
	sloggin "github.com/samber/slog-gin"
)

type Handler struct {
	services   *service.Service
	logger     *slog.Logger
	middleware *middleware.AuthInterceptor
}

func NewHandler(services *service.Service, logger *slog.Logger) *Handler {
	return &Handler{
		services:   services,
		logger:     logger,
		middleware: middleware.NewAuthInterceptor(services.Authorization),
	}
}

func (h *Handler) InitRoutes(env string, allowOrigins []string) *gin.Engine {
	// Set output mode
	if env != "local" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize default router
	router := gin.Default()

	// Cap the memory used to buffer multipart uploads before they spill to disk.
	router.MaxMultipartMemory = 16 << 20 // 16 MB

	// Router settings
	router.Use(gin.Recovery())        // recovery middleware for panics
	router.Use(sloggin.New(h.logger)) // use custom logger for output messages
	router.Use(cors.New(cors.Config{
		AllowOrigins:     append(allowOrigins, "http://localhost:3000"),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})) // CORS Policy

	// Registry new validation rules for struct tags in domain models or sessions
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("city", valid.ValidCity); err != nil {
			panic("failed to registry new validation rule for city: " + err.Error())
		}
	}

	auth := router.Group("/auth")
	{
		auth.POST("/signup", h.signUp)
		auth.POST("/login", h.signIn)
		auth.POST("/refresh", h.refresh)
	}

	api := router.Group("/api")
	{
		api.POST("/logout", h.middleware.UserIdentity, h.logout)

		vacancies := api.Group("/vacancies")
		{
			vacancies.GET("/", h.getAllVacancies)
			vacancies.GET("/:id", h.getVacancy)
			vacancies.POST("/respond", h.respond)
		}

		plan := api.Group("/plan")
		{
			plan.POST("/", h.requestPlan)
			plan.GET("/:id", h.middleware.UserIdentity, h.getResponseFromRequest)
		}

		admin := api.Group("/admin", h.middleware.UserIdentity, h.middleware.AdminIdentity)
		{
			vacanciesAdmin := admin.Group("/vacancies")
			{
				vacanciesAdmin.POST("/", h.createVacancy)
				vacanciesAdmin.PUT("/:id", h.updateVacancy)
				vacanciesAdmin.DELETE("/:id", h.deleteVacancy)

				vacanciesAdmin.GET("/", h.getRespondVacancies)
				vacanciesAdmin.GET("/:id", h.getRespondVacancy)
			}

			planAdmin := admin.Group("/plans")
			{
				planAdmin.GET("/", h.getAllPlans)
				planAdmin.GET("/:id", h.getPlan)
			}
		}
	}

	return router
}
