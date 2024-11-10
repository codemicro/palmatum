package httpsrv

import (
	"github.com/codemicro/palmatum/palmatum/internal/config"
	"github.com/codemicro/palmatum/palmatum/internal/core"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
)

func New(conf *config.Config, c *core.Core) (http.Handler, error) {
	r := &routes{
		config: conf,
		core:   c,
	}

	router := httprouter.New()

	router.GET("/-/", r.managementIndex)
	router.POST("/-/upload", r.uploadSite)
	router.POST("/-/delete", r.deleteSite)

	return router, nil
}

type routes struct {
	config *config.Config
	core   *core.Core
}

func BadRequestResponse(w http.ResponseWriter, message ...string) error {
	outputMessage := "Bad Request"
	if len(message) != 0 {
		outputMessage = message[0]
	}
	w.WriteHeader(400)
	_, err := w.Write([]byte(outputMessage))
	return err
}

func IsBrowser(r *http.Request) bool {
	if r.Header.Get("HX-Request") != "" {
		return true
	}
	sp := strings.Split(r.Header.Get("Accept"), ",")
	for _, item := range sp {
		if item == "" {
			continue
		}
		x := strings.Split(item, ";")
		if strings.EqualFold(x[0], "text/html") {
			return true
		}
	}
	return false
}
