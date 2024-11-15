package httpsrv

import (
	"net/http"
)

func NewSitesServer(args ServerArgs) *http.Server {
	return newServer(args, args.Config.HTTP.SitesAddress, http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
		args.Core.RouteRequest(rq).ServeHTTP(rw, rq)
	}))
}
