package httpsrv

import (
	"net/http"
)

func NewSitesServer(args ServerArgs) *http.Server {
	// TODO: serve sites
	return newServer(args, args.Config.HTTP.SitesAddress, New(args.Config, args.Core))
}
