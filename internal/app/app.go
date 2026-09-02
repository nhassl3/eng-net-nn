package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/nhassl3/IpBuild-backend/internal/config"
	"github.com/nhassl3/IpBuild-backend/internal/db"
	postgres2 "github.com/nhassl3/IpBuild-backend/internal/repository/postgres"
	redis3 "github.com/nhassl3/IpBuild-backend/internal/repository/redis"
	"github.com/nhassl3/IpBuild-backend/internal/service"
	handle "github.com/nhassl3/IpBuild-backend/internal/transport/gin-http"
	"github.com/nhassl3/IpBuild-backend/pkg/auth"
	"github.com/nhassl3/IpBuild-backend/pkg/logger"
	"github.com/nhassl3/IpBuild-backend/pkg/mailer"
	"github.com/nhassl3/IpBuild-backend/pkg/minio"
	"github.com/nhassl3/IpBuild-backend/pkg/postgres"
	redis2 "github.com/nhassl3/IpBuild-backend/pkg/redis"
)

type Server struct {
	httpServer *http.Server
	notifier   mailer.Notifier
}

func (s *Server) Run(cfg *config.Config, log logger.Logger) error {
	ctx := context.Background()

	dsn := postgres.DSN(
		cfg.DBSettings.Host,
		cfg.DBSettings.Port,
		cfg.DBSettings.Username,
		cfg.DBSettings.Password,
		cfg.DBSettings.DBName,
		cfg.DBSettings.SSLMode,
	)
	pool, err := postgres.NewPool(ctx, dsn, log.Named("postgres"))
	if err != nil {
		return fmt.Errorf("app: connect postgres: %w", err)
	}
	defer pool.Close()

	log.Info("connected to postgres")

	store := db.NewStore(pool)
	repo := postgres2.NewRepository(store)

	redisClient, err := redis2.NewRedisClient(ctx,
		cfg.RedisServer.Address,
		cfg.RedisServer.Username,
		cfg.RedisServer.Password,
		cfg.RedisServer.DB,
		log.Named("redis"),
	)
	if err != nil {
		return fmt.Errorf("app: connect redis: %w", err)
	}

	log.Info("connected to redis")

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
			cfg.SMTP.From, cfg.SMTP.WorkEmail, cfg.SMTP.Port, log.Named("mailer"),
		)
		if err != nil {
			return fmt.Errorf("app: create SMTP mailer: %w", err)
		}
	} else {
		notifier = mailer.NewNoopNotifier(log.Named("nop_mailer"))
	}
	s.notifier = notifier

	minIOClient, err := minio.NewMinIO(
		ctx,
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKey,
		cfg.MinIO.SecretKey,
		"",
		cfg.MinIO.Bucket,
		cfg.UseSSL,
	)
	if err != nil {
		return fmt.Errorf("app: create minio client: %w", err)
	}
	log.Info("connected to minio")

	services := service.NewService(
		repo,
		authRedis,
		accessManagerWithBL,
		refreshManagerWithBL,
		blacklistRepo,
		notifier,
		minIOClient,
		log,
	)

	handler := handle.NewHandler(services, log.Named("http"), &cfg.Token)

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
	if s.httpServer == nil {
		return nil
	}
	httpErr := s.httpServer.Shutdown(ctx)
	mailerErr := s.notifier.Close(ctx)
	return errors.Join(httpErr, mailerErr)
}
