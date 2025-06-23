package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/tarantool/go-tarantool/v2"
	"github.com/tmybsv/tarantool-kv/internal/storage"
	"github.com/tmybsv/tarantool-kv/internal/transport/http/handler"
	"github.com/tmybsv/tarantool-kv/internal/transport/http/middleware"
)

// App is an initialized application.
type App struct {
	HTTPServer *http.Server
	conn       *tarantool.Connection
	opts       Options
}

// Options is the application options.
type Options struct {
	TarantoolAddr     string
	TarantoolUser     string
	TarantoolPassword string
	TarantoolTimeout  time.Duration
	TarantoolKVSpace  string
	TarantoolKVIndex  string
	HTTPKVBasePath    string
	HTTPAddr          string
	HTTPTimeout       time.Duration
}

// New creates a new application.
func New(log *slog.Logger, ctx context.Context, opts Options) (*App, error) {
	tarantoolDialer := tarantool.NetDialer{
		Address:  opts.TarantoolAddr,
		User:     opts.TarantoolUser,
		Password: opts.TarantoolPassword,
	}
	tarantoolOpts := tarantool.Opts{
		Timeout: opts.TarantoolTimeout,
	}

	tarantoolConn, err := tarantool.Connect(ctx, tarantoolDialer, tarantoolOpts)
	if err != nil {
		return nil, fmt.Errorf("connect to Tarantool: %w", err)
	}

	ts := storage.NewTarantool(tarantoolConn, opts.TarantoolKVSpace, opts.TarantoolKVIndex)
	kvHandler := handler.NewKV(log, ts, opts.HTTPKVBasePath)

	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodPost, opts.HTTPKVBasePath), kvHandler.Set)
	mux.HandleFunc(fmt.Sprintf("%s %s/{key}", http.MethodGet, opts.HTTPKVBasePath), kvHandler.Get)
	mux.HandleFunc(fmt.Sprintf("%s %s/{key}", http.MethodPut, opts.HTTPKVBasePath), kvHandler.Update)
	mux.HandleFunc(fmt.Sprintf("%s %s/{key}", http.MethodDelete, opts.HTTPKVBasePath), kvHandler.Delete)
	loggingMiddleware := middleware.Logging(log, mux)

	server := &http.Server{
		Addr:         opts.HTTPAddr,
		Handler:      loggingMiddleware,
		ReadTimeout:  opts.HTTPTimeout,
		WriteTimeout: opts.HTTPTimeout,
	}

	return &App{
		HTTPServer: server,
		conn:       tarantoolConn,
	}, nil
}

// Stop stops the application.
func (a *App) Stop(ctx context.Context) {
	a.conn.Close()
	a.HTTPServer.Shutdown(ctx)
}
