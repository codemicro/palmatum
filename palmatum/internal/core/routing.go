package core

import (
	"fmt"
	"github.com/codemicro/palmatum/palmatum/internal/caddyController"
	"net/http"
	"os"
	"strings"
)

type routeDestination struct {
	ID          int    `db:"id"`
	Domain      string `db:"domain"`
	Path        string `db:"path"`
	ContentPath string `db:"content_path"`
}

type cachedHandler struct {
	Handler http.Handler
	file    *os.File
}

func (c *cachedHandler) Close() error {
	return c.file.Close()
}

func (c *Core) BuildKnownRoutes() error {
	// TODO: tidy up this function, adapt the logging to make sense with the new Caddy setup and rename it

	c.Logger.Debug("loading known routes into memory")
	c.routeLock.Lock()
	defer c.routeLock.Unlock()

	var destinations []*caddyController.RouteDestination
	if err := c.Database.Select(&destinations, `SELECT routes.id, routes.domain, routes.path, sites.content_path FROM routes JOIN sites ON routes.site = sites.slug;`); err != nil {
		return fmt.Errorf("read from database: %w", err)
	}

	kr := make(caddyController.RouteSpec)

	for _, d := range destinations {
		c.Logger.Debug("registering route", "domain", d.Domain, "path", d.Path, "id", d.ID)
		d.Domain = strings.ToLower(d.Domain)
		kr[d.Domain] = append(kr[d.Domain], d)
	}

	if err := c.CaddyController.Reconfigure(kr); err != nil {
		return fmt.Errorf("reconfigure Caddy controller: %w", err)
	}
	return nil
}
