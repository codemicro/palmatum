package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/codemicro/palmatum/palmatum/internal/config"
	"github.com/codemicro/palmatum/palmatum/internal/datastore"
	"github.com/codemicro/palmatum/palmatum/internal/httpsrv"
	"golang.org/x/exp/slog"
)

func main() {
	if err := run(); err != nil {
		return
	}
}

func run() error {
	conf, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config on startup: %w", err)
	}

	_ = os.MkdirAll(conf.Platform.SitesDirectory, 0777)

	ds := datastore.New(conf.Database.StoreFilename)

	handler, err := httpsrv.New(conf, ds)
	if err != nil {
		return fmt.Errorf("creating HTTP handler: %w", err)
	}

	host := fmt.Sprintf("%s:%d", conf.HTTP.Host, conf.HTTP.Port)
	slog.Info("http server alive", "host", host)

	err = http.ListenAndServe(host, handler)
	if err != nil {
		return fmt.Errorf("serving HTTP: %w", err)
	}

	return nil
}
