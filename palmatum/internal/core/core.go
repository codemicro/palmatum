package core

import (
	"context"
	"github.com/codemicro/palmatum/palmatum/internal/caddyController"
	"github.com/codemicro/palmatum/palmatum/internal/config"
	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"
	"log/slog"
	"path"
	"sync"
)

type Core struct {
	Config          *config.Config
	Database        *sqlx.DB
	Logger          *slog.Logger
	CaddyController *caddyController.Controller

	routeLock   sync.RWMutex
	knownRoutes map[string][]*routeDestination

	handlerCacheLock sync.Mutex
	handlerCache     map[string]*cachedHandler
}

func New(lc fx.Lifecycle, c *config.Config, db *sqlx.DB, logger *slog.Logger, cctrl *caddyController.Controller) *Core {
	co := &Core{
		Config:          c,
		Database:        db,
		Logger:          logger,
		CaddyController: cctrl,
	}

	lc.Append(fx.Hook{OnStart: func(ctx context.Context) error {
		return co.BuildKnownRoutes()
	}})

	return co
}

func (c *Core) getPathOnDisk(p string) string {
	return path.Join(c.Config.Platform.SitesDirectory, p)
}
