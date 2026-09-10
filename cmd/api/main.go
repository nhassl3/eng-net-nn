package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nhassl3/IpBuild-backend/internal/app"
	"github.com/nhassl3/IpBuild-backend/internal/config"
	"github.com/nhassl3/IpBuild-backend/pkg/logger"
)

// version is the build version, injected via -ldflags "-X main.version=...".
// Left as "dev" for local/unversioned builds.
var version = "dev"

func main() {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		env := os.Getenv("ENVIRONMENT")
		switch env {
		case "prod":
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
		// The logger isn't built yet (it depends on cfg.Log), so this one
		// failure path goes to stderr directly instead of through it.
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.New(logger.Config{
		Level:      cfg.Log.Level,
		AddCaller:  cfg.Log.AddCaller,
		Stacktrace: cfg.Log.Stacktrace,
		Env:        cfg.Env,
		Service:    "ipbuild-backend",
		Version:    version,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = log.Sync() }()

	log = log.Named("api")
	log.Info("starting server")

	server := new(app.Server)
	if err := server.New(cfg, log); err != nil {
		log.Error("error initializing server", logger.Err(err))
		os.Exit(1)
	}
	go func() {
		if err := server.Run(); err != nil {
			log.Error("error starting server", logger.Err(err))
		}
	}()
	log.Info("server started", logger.String("address", cfg.HttpServer.Address))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("server is shutting down")
	shutdownStart := time.Now()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("error shutting down the server", logger.Err(err))
	}
	log.Info("shutdown complete", logger.Duration(time.Since(shutdownStart)))
}
