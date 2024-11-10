package core

import (
	"github.com/codemicro/palmatum/palmatum/internal/config"
	"github.com/jmoiron/sqlx"
	"path"
)

type Core struct {
	Config   *config.Config
	Database *sqlx.DB
}

func New(c *config.Config, db *sqlx.DB) *Core {
	return &Core{
		Config:   c,
		Database: db,
	}
}

func (c *Core) getPathOnDisk(p string) string {
	return path.Join(c.Config.Platform.SitesDirectory, p)
}
