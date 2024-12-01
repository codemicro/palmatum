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
