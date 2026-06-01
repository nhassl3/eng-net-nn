package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/nhassl3/IpBuild-backend/internal/config"
	"github.com/nhassl3/IpBuild-backend/internal/db"
	postgres2 "github.com/nhassl3/IpBuild-backend/internal/repository/postgres"
	redis3 "github.com/nhassl3/IpBuild-backend/internal/repository/redis"
	"github.com/nhassl3/IpBuild-backend/internal/service"
	handle "github.com/nhassl3/IpBuild-backend/internal/transport/gin-http"
	"github.com/nhassl3/IpBuild-backend/pkg/auth"
	"github.com/nhassl3/IpBuild-backend/pkg/mailer"
	"github.com/nhassl3/IpBuild-backend/pkg/postgres"
	redis2 "github.com/nhassl3/IpBuild-backend/pkg/redis"
)

type Server struct {
	httpServer *http.Server
	notifier   mailer.Notifier
}

func (s *Server) Run(cfg *config.Config, logger *slog.Logger) error {
	ctx := context.Background()

	logger.Info("Allow-Origins", slog.Any("allow_origins", cfg.AllowOrigins))

	dsn := postgres.DSN(
		cfg.DBSettings.Host,
		cfg.DBSettings.Port,
		cfg.DBSettings.Username,
		cfg.DBSettings.Password,
		cfg.DBSettings.DBName,
		cfg.DBSettings.SSLMode,
	)
	pool, err := postgres.NewPool(ctx, dsn)
	if err != nil {
		return fmt.Errorf("app: connect postgres: %w", err)
	}
	defer pool.Close()

	store := db.NewStore(pool)
	repo := postgres2.NewRepository(store)

	redisClient, err := redis2.NewRedisClient(ctx,
		cfg.RedisServer.Address,
		cfg.RedisServer.Username,
		cfg.RedisServer.Password,
		cfg.RedisServer.DB,
	)
	if err != nil {
		return fmt.Errorf("app: connect redis: %w", err)
	}

	authRedis := redis3.NewAuthRedisRepository(redisClient, cfg.RedisServer.TTL.UserProfile)
	blacklistRepo := redis3.NewBlacklistRepository(redisClient)

	accessMaker, err := auth.NewPASETOMaker(cfg.Token.PasetoKeyHex, cfg.Token.AccessTTL)
	if err != nil {
		return fmt.Errorf("app: create access token maker: %w", err)
	}
	refreshMaker, err := auth.NewPASETOMaker(cfg.Token.PasetoKeyHex, cfg.Token.RefreshTTL)
	if err != nil {
		return fmt.Errorf("app: create refresh token maker: %w", err)
	}

	accessManagerWithBL := auth.NewBlacklistedTokenManager(accessMaker, blacklistRepo)
	refreshManagerWithBL := auth.NewBlacklistedTokenManager(refreshMaker, blacklistRepo)

	var notifier mailer.Notifier
	if cfg.SMTP.Host != "" {
		notifier, err = mailer.NewSMTPMailer(
			cfg.SMTP.Host, cfg.SMTP.Username, cfg.SMTP.Password,
			cfg.SMTP.From, cfg.SMTP.WorkEmail, cfg.SMTP.Port, logger,
		)
		if err != nil {
			return fmt.Errorf("app: create SMTP mailer: %w", err)
		}
	} else {
		notifier = &mailer.NoopNotifier{}
	}
	s.notifier = notifier

	services := service.NewService(repo, authRedis, accessManagerWithBL, refreshManagerWithBL, blacklistRepo, notifier)

	handler := handle.NewHandler(services, logger)

	s.httpServer = &http.Server{
		Addr:           cfg.HttpServer.Address,
		MaxHeaderBytes: 1 << 20,
		Handler:        handler.InitRoutes(cfg.Env, cfg.AllowOrigins),
		ReadTimeout:    cfg.HttpServer.Timeout,
		WriteTimeout:   cfg.HttpServer.Timeout,
		IdleTimeout:    cfg.HttpServer.IdleTimeout,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	httpErr := s.httpServer.Shutdown(ctx)
	mailerErr := s.notifier.Close(ctx)
	return errors.Join(httpErr, mailerErr)
}
