package gin_http

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/nhassl3/IpBuild-backend/internal/config"
	"github.com/nhassl3/IpBuild-backend/internal/service"
	"github.com/nhassl3/IpBuild-backend/internal/transport/gin-http/middleware"
	valid "github.com/nhassl3/IpBuild-backend/internal/transport/gin-http/validator"
	"github.com/nhassl3/IpBuild-backend/pkg/logger"
)

type Handler struct {
	services   *service.Service
	logger     logger.Logger
	middleware *middleware.AuthInterceptor
	tokenCfg   *config.Token
}

func NewHandler(services *service.Service, logger logger.Logger, tokenCfg *config.Token) *Handler {
	return &Handler{
		services:   services,
		logger:     logger,
		middleware: middleware.NewAuthInterceptor(services.Authorization),
		tokenCfg:   tokenCfg,
	}
}

func (h *Handler) InitRoutes(env string, allowOrigins []string) *gin.Engine {
	// Set output mode
	if env != "local" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Bare engine: gin.Default() bundles its own Logger/Recovery, which we
	// replace with structured equivalents below.
	router := gin.New()

	// Cap the memory used to buffer multipart uploads before they spill to disk.
	router.MaxMultipartMemory = 16 << 20 // 16 MB

	// Router settings
	router.Use(middleware.RequestID())       // assigns/propagates X-Request-ID
	router.Use(middleware.Logging(h.logger)) // structured access log + request-scoped logger in ctx
	router.Use(middleware.Recovery())        // panic recovery with stacktrace, must run after Logging
	// One list for both CORS and the cross-site guard, so an origin can never
	// be allowed by one and rejected by the other.
	origins := append(append([]string{}, allowOrigins...), "http://localhost:3000")

	router.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
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

	// The refresh cookie is SameSite=None in production, so every endpoint that
	// acts on it needs CSRF protection: an origin allowlist on the group, plus
	// a preflight-forcing header on /refresh, which authenticates by cookie alone.
	auth := router.Group("/auth", middleware.CrossSiteGuard(origins))
	{
		auth.POST("/signup", h.signUp)
		auth.POST("/login", h.signIn)
		auth.POST("/refresh", middleware.RequireRequestedWith, h.refresh)
	}

	api := router.Group("/api")
	{
		api.POST("/logout", middleware.CrossSiteGuard(origins), h.middleware.UserIdentity, h.logout)
		api.GET("/me", h.middleware.UserIdentity, h.me)

		vacancies := api.Group("/vacancies")
		{
			vacancies.GET("", h.getAllVacancies)
			vacancies.GET("/:id", h.getVacancy)
			vacancies.POST("/respond", h.respond)
		}

		plan := api.Group("/plans")
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

			jobDirectionsAdmin := admin.Group("/job_directions")
			{
				jobDirectionsAdmin.GET("", h.listJd)
				jobDirectionsAdmin.GET("/:id", h.getJd)

				jobDirectionsAdmin.POST("/", h.createJd)
				jobDirectionsAdmin.PUT("/:id", h.updateJd)
				jobDirectionsAdmin.DELETE("/:id", h.deleteJd)
			}

			planAdmin := admin.Group("/plans")
			{
				planAdmin.GET("/", h.getAllPlans)
				planAdmin.GET("/:id", h.getPlan)
				// TODO: POST response to plan
			}
		}
	}

	return router
}
