package httpsrv

import (
	"github.com/codemicro/rubrum/rubrum/internal/config"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
)

func New(conf *config.Config) (http.Handler, error) {
	r := &routes{
		config: conf,
	}

	router := httprouter.New()

	router.GET("/-/", r.managementIndex)
	router.POST("/-/upload", r.uploadSite)
	router.POST("/-/delete", r.deleteSite)

	return router, nil
}

type routes struct {
	config *config.Config
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
