package httpsrv

import (
	"fmt"
	"net/http"
)

func NewSitesServer(args ServerArgs) *http.Server {
	return newServer(args, args.Config.HTTP.SitesAddress, handleErrors(args.Logger, func(rw http.ResponseWriter, rq *http.Request) error {
		h, err := args.Core.RouteRequest(rq)
		if err != nil {
			return fmt.Errorf("handle sites request: %w", err)
		}
		h.ServeHTTP(rw, rq)
		return nil
	}))
}
