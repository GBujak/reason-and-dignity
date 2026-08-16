package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/gbujak/reason-and-dignity/m/v2/internal/app"
)

type Options struct {
	Port int `help:"Port to listen on" short:"p" default:"8888"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cli := humacli.New(func(h humacli.Hooks, o *Options) {
		_, mux := app.NewApi()

		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", o.Port),
			Handler: mux,
		}

		h.OnStart(func() {
			slog.Info("Starting server", "port", o.Port)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("Server failed", "error", err)
			}
		})

		h.OnStop(func() {
			slog.Info("Stopping server...")
			server.Shutdown(context.Background())
		})
	})

	cli.Run()
}
