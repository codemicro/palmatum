package httpsrv

import (
	"context"
	"github.com/codemicro/palmatum/palmatum/internal/config"
	"github.com/codemicro/palmatum/palmatum/internal/core"
	"go.uber.org/fx"
	"log/slog"
	"net/http"
	"strings"
)

type ServerArgs struct {
	fx.In

	Lifecycle  fx.Lifecycle
	Shutdowner fx.Shutdowner
	Logger     *slog.Logger
	Config     *config.Config
	Core       *core.Core
}

func newServer(args ServerArgs, addr string, handler http.Handler) *http.Server {
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	args.Lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			args.Logger.Info("http server alive", "address", addr)
			go func() {
				if err := server.ListenAndServe(); err != nil {
					args.Logger.Error("failed to start HTTP server", "address", addr, "error", err)
					_ = args.Shutdowner.Shutdown(fx.ExitCode(2))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})

	return server
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
