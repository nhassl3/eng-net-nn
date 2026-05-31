package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nhassl3/IpBuild-backend/internal/app"
	"github.com/nhassl3/IpBuild-backend/internal/config"
	"github.com/nhassl3/IpBuild-backend/pkg/logger/handlers/slogpretty"
	"github.com/nhassl3/IpBuild-backend/pkg/logger/sl"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		env := os.Getenv("ENVIRONMENT")
		switch env {
		case envProd:
			configFile = "config/prod.yaml"
		default:
			configFile = "config/local.yaml"
		}
	}

	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}

	cfg, err := config.Load(configFile, envFile)
	if err != nil {
		slog.Error("failed to load config", sl.ErrLog(err))
		return
	}

	log := setupLogger(cfg.Env)
	log.Info("starting server", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	server := new(app.Server)
	go func() {
		if err := server.Run(cfg, log); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("error starting server", sl.ErrLog(err))
		}
	}()
	log.Info("Server started", slog.String("host", cfg.HttpServer.Address))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Server is down")
	if err := server.Shutdown(context.Background()); err != nil {
		log.Error("error shutting down the server", sl.ErrLog(err))
	}
}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = setupPrettySlogger(slog.LevelDebug)
	case envProd:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))
	default:
		logger = setupPrettySlogger()
	}

	return logger
}

func setupPrettySlogger(level ...slog.Level) *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: level[0],
		},
	}
	return slog.New(opts.NewPrettyHandler(os.Stdout))
}
