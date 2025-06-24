package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/tmybsv/tarantool-kv/internal/app"
	"github.com/tmybsv/tarantool-kv/internal/config"
)

const (
	envProduction  = "prod"
	envDevelopment = "dev"
	envLocal       = "local"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info("starting application", slog.String("env", cfg.Env))
	app, err := app.New(log, ctx, app.Options{
		TarantoolAddr:     fmt.Sprintf("%s:%d", cfg.Tarantool.Host, cfg.Tarantool.Port),
		TarantoolUser:     cfg.Tarantool.User,
		TarantoolPassword: cfg.Tarantool.Password,
		TarantoolTimeout:  cfg.Tarantool.Timeout,
		TarantoolKVSpace:  cfg.Tarantool.KVSpace,
		TarantoolKVIndex:  cfg.Tarantool.KVIndex,
		HTTPKVBasePath:    cfg.HTTP.KVBasePath,
		HTTPAddr:          fmt.Sprintf(":%d", cfg.HTTP.Port),
		HTTPTimeout:       cfg.HTTP.Timeout,
	})
	if err != nil {
		log.Error("failed to init application", slog.String("error", err.Error()))
		os.Exit(1)
	}

	go func() {
		log.Info("HTTP server is starting", slog.Int("port", cfg.HTTP.Port))
		if err := app.HTTPServer.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Error("failed to start HTTP server", slog.String("error", err.Error()))
				os.Exit(1)
			}
			log.Info("HTTP server stopped")
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	<-shutdown
	log.Info("gracefully shutting down")
	app.Stop(ctx)
}

func setupLogger(env string) *slog.Logger {
	log := &slog.Logger{}
	switch env {
	case envProduction:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	case envDevelopment:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	case envLocal:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		}))
	}
	return log
}
