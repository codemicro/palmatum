package core

import (
	"archive/zip"
	"fmt"
	"net/http"
	"os"
	"slices"
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
	c.Logger.Debug("loading known routes into memory")
	c.routeLock.Lock()
	defer c.routeLock.Unlock()

	var destinations []*routeDestination
	if err := c.Database.Select(&destinations, `SELECT routes.id, routes.domain, routes.path, sites.content_path FROM routes JOIN sites ON routes.site = sites.slug;`); err != nil {
		return fmt.Errorf("read from database: %w", err)
	}

	kr := make(map[string][]*routeDestination)

	for _, d := range destinations {
		c.Logger.Debug("registering route", "domain", d.Domain, "path", d.Path, "id", d.ID)
		d.Domain = strings.ToLower(d.Domain)
		kr[d.Domain] = append(kr[d.Domain], d)
	}

	for _, v := range kr {
		// sort longest path first and hence match longest path first
		slices.SortFunc(v, func(a, b *routeDestination) int {
			sa := len(strings.Split(a.Path, "/"))
			sb := len(strings.Split(b.Path, "/"))

			if sa < sb {
				return 1
			} else if sa > sb {
				return -1
			}
			return 0
		})
	}

	c.knownRoutes = kr

	c.handlerCacheLock.Lock()
	defer c.handlerCacheLock.Unlock()

	if c.handlerCache != nil {
		c.Logger.Debug("closing existing cached handlers")
		for k, v := range c.handlerCache {
			if err := v.Close(); err != nil {
				c.Logger.Warn("failed to close cached handler", "error", err, "content", k)
			}
		}
	}

	c.handlerCache = make(map[string]*cachedHandler)

	return nil
}

var (
	invalidRequestHandler = makeErrorHandler(http.StatusBadRequest, "Invalid request")
	notFoundHandler       = makeErrorHandler(http.StatusNotFound, "No maching route found")
)

func makeErrorHandler(statusCode int, message string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(statusCode)
		_, _ = rw.Write([]byte(message))
	})
}

func (c *Core) RouteRequest(rq *http.Request) (http.Handler, error) {
	c.routeLock.RLock()
	defer c.routeLock.RUnlock()

	host := strings.ToLower(rq.Host)

	if host == "" {
		return invalidRequestHandler, nil
	}

	c.Logger.Debug("routing request", "host", host, "path", rq.URL.Path, "foundRoutes", c.knownRoutes[host])

	for _, v := range c.knownRoutes[host] {
		if strings.HasPrefix(rq.URL.Path, v.Path) {
			p := c.getPathOnDisk(v.ContentPath)

			c.handlerCacheLock.Lock()

			if h, found := c.handlerCache[v.ContentPath]; found {
				c.Logger.Debug("handler cache HIT", "content", v.ContentPath)
				c.handlerCacheLock.Unlock()
				return h.Handler, nil
			}
			c.Logger.Debug("handler cache MISS", "content", v.ContentPath)

			fi, err := os.Stat(p)
			if err != nil {
				return nil, fmt.Errorf("stat content file: %w", err)
			}

			f, err := os.Open(p)
			if err != nil {
				return nil, fmt.Errorf("open content file: %w", err)
			}

			zr, err := zip.NewReader(f, fi.Size())
			if err != nil {
				return nil, fmt.Errorf("create ZIP reader: %w", err)
			}

			fs := http.FileServerFS(zr)

			handler := http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
				rq.URL.Path = rq.URL.Path[len(v.Path):]
				if rq.URL.Path == "" {
					rq.URL.Path = "/"
				}

				fs.ServeHTTP(rw, rq)
			})

			c.handlerCache[v.ContentPath] = &cachedHandler{
				Handler: handler,
				file:    f,
			}

			c.handlerCacheLock.Unlock()

			return handler, nil
		}
	}

	return notFoundHandler, nil
}
