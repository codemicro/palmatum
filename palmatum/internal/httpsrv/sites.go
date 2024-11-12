package httpsrv

import (
	"fmt"
	"net/http"
)

func NewSitesServer(args ServerArgs) *http.Server {
	return newServer(args, fmt.Sprintf("%s:%d", args.Config.HTTP.Host, args.Config.HTTP.Port+2), New(args.Config, args.Core))
}
