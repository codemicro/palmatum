package caddyController

import (
	"slices"
	"strings"
)

type RouteDestination struct {
	ID          int    `db:"id"`
	Domain      string `db:"domain"`
	Path        string `db:"path"`
	ContentPath string `db:"content_path"`
}

// RouteSpec maps domains to a set of routes within them and describes how to map them all together
type RouteSpec map[string][]*RouteDestination

func (rs RouteSpec) sortValues() {
	for _, v := range rs {
		// sort longest path first and hence match longest path first
		slices.SortFunc(v, func(a, b *RouteDestination) int {
			sa := len(strings.Split(strings.TrimRight(a.Path, "/"), "/"))
			sb := len(strings.Split(strings.TrimRight(b.Path, "/"), "/"))

			if sa < sb {
				return 1
			} else if sa > sb {
				return -1
			}
			return 0
		})
	}
}

type caddyConfig struct {
	// Apps is intended to be filled with caddyFilesystemApp and caddyHttpApp
	Apps map[string]any `json:"apps"`
}

type caddyFilesystem struct {
	Name       string `json:"name"`
	Filesystem struct {
		Backend string `json:"backend"`
		Path    string `json:"path"`
	} `json:"file_system"`
}

type caddyFilesystemApp struct {
	Filesystems []*caddyFilesystem `json:"filesystems"`
}

type caddyHttpApp struct {
	HttpPort int                         `json:"http_port"`
	Servers  map[string]*caddyHttpServer `json:"servers"`
}

type caddyHttpServer struct {
	Listen         []string          `json:"listen"`
	Routes         []*caddyBaseRoute `json:"routes"`
	AutomaticHttps struct {
		Disable bool `json:"disable"`
	} `json:"automatic_https"`
}

type caddyMatcher struct {
	Host []string `json:"host,omitempty"`
	Path []string `json:"path,omitempty"`
}

type caddyFileserverHandler struct {
	Fs string `json:"fs"`

	// Handler must always be set to file_server
	Handler string `json:"handler"`
}

type caddyRoute struct {
	Handle []*caddyFileserverHandler `json:"handle"`
	Match  []*caddyMatcher           `json:"match"`
}

type caddySubroute struct {
	// Handler should always be set to subroute
	Handler string        `json:"handler"`
	Routes  []*caddyRoute `json:"routes"`
}

type caddyBaseRoute struct {
	Match    []*caddyMatcher  `json:"match"`
	Handle   []*caddySubroute `json:"handle"`
	Terminal bool             `json:"terminal"`
}
