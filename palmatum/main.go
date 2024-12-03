package main

import (
	"github.com/codemicro/palmatum/palmatum/internal/caddyController"
	"github.com/codemicro/palmatum/palmatum/internal/config"
	"github.com/codemicro/palmatum/palmatum/internal/core"
	"github.com/codemicro/palmatum/palmatum/internal/database"
	"github.com/codemicro/palmatum/palmatum/internal/httpsrv"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	fx.New(
		fx.Provide(provideLogger),
		fx.WithLogger(provideFxLogger),

		fx.Provide(
			config.Load,
			database.New,
			caddyController.NewController,
			core.New,
			httpsrv.NewManagementServer,
		),

		fx.Invoke(
			func(conf *config.Config) {
				_ = os.MkdirAll(conf.Platform.SitesDirectory, 0777)
			},
			func(log *slog.Logger, _ *http.Server, _ *caddyController.Controller) {
				log.Info("Palmatum has started")
			},
		),
	).Run()
}

func provideLogger(conf *config.Config) *slog.Logger {
	level := new(slog.LevelVar)
	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	if conf.Debug {
		level.Set(slog.LevelDebug)
		l.Debug("debug mode enabled")
	}

	return l
}

func provideFxLogger(l *slog.Logger) fxevent.Logger {
	fxel := &fxevent.SlogLogger{
		Logger: l.With("area", "fx"),
	}
	fxel.UseLogLevel(slog.LevelDebug)
	return fxel
}
