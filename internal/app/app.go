package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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
	"github.com/redis/go-redis/v9"
)

type Server struct {
	httpServer  *http.Server
	notifier    mailer.Notifier
	pool        *pgxpool.Pool
	redisClient *redis.Client
}

// New synchronously builds all the server's dependencies and the
// underlying http.Server. It must complete before Run or Shutdown are
// called from other goroutines, so that s.httpServer and friends are fully
// initialized before they are ever read concurrently.
func (s *Server) New(cfg *config.Config, log logger.Logger) error {
	// Context initialize
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Postgres initialize
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
	s.pool = pool

	log.Info("connected to postgres")

	// Postgres store initialize
	store := db.NewStore(pool)
	repo := postgres2.NewRepository(store)

	// Redis initialize
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
	s.redisClient = redisClient

	log.Info("connected to redis")

	// Redis store initialize
	authRedis := redis3.NewAuthRedisRepository(redisClient, cfg.RedisServer.TTL.UserProfile)
	blacklistRepo := redis3.NewBlacklistRepository(redisClient)

	// Access and refresh with blacklist maker initialize
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

	// Notifier initialize
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

	// MinIO initialize
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

	// Services initialize
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

	// Initialize handler
	handler := handle.NewHandler(services, log.Named("http"), &cfg.Token)

	// Formate http server
	s.httpServer = &http.Server{
		Addr:           cfg.HttpServer.Address,
		MaxHeaderBytes: 1 << 20,
		Handler:        handler.InitRoutes(cfg.Env, cfg.AllowOrigins),
		ReadTimeout:    cfg.HttpServer.Timeout,
		WriteTimeout:   cfg.HttpServer.Timeout,
		IdleTimeout:    cfg.HttpServer.IdleTimeout,
	}

	return nil
}

// Run blocks serving HTTP on the http.Server built by New, until the
// listener fails or Shutdown closes it. New must be called first.
func (s *Server) Run() error {
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("app: Run: failed to start http server: %w", err)
	}
	return nil
}

// Shutdown stops the HTTP server first, then closes the remaining
// dependencies in the reverse order they were initialized in Run.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	httpErr := s.httpServer.Shutdown(ctx)

	notifierErr := s.notifier.Close(ctx)

	var redisErr error
	if s.redisClient != nil {
		redisErr = s.redisClient.Close()
	}
	if s.pool != nil {
		s.pool.Close()
	}

	return errors.Join(httpErr, notifierErr, redisErr)
}
