package httpsrv

import (
	"net/http"
)

func NewSitesServer(args ServerArgs) *http.Server {
	// TODO: serve sites
	return newServer(args, args.Config.HTTP.SitesAddress, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("not implemented"))
	}))
}
