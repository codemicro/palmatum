package caddyController

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"git.tdpain.net/codemicro/palmatum/palmatum/internal/config"
	"go.uber.org/fx"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Controller struct {
	logger *slog.Logger
	config *config.Config

	cmd            *exec.Cmd
	adminApiSocket string
}

func NewController(lc fx.Lifecycle, logger *slog.Logger, conf *config.Config) *Controller {
	csc := &Controller{
		logger:         logger,
		config:         conf,
		adminApiSocket: "localhost:52019",
	}

	csc.cmd = exec.Command(conf.Platform.CaddyExecutablePath, "run")
	csc.cmd.Env = append(csc.cmd.Env, "CADDY_ADMIN="+csc.adminApiSocket)
	csc.cmd.Stdout = os.Stdout
	csc.cmd.Stderr = os.Stderr

	lc.Append(fx.Hook{
		OnStart: csc.start,
		OnStop:  csc.stop,
	})

	return csc
}

func (csc *Controller) start(context.Context) error {
	csc.logger.Info("starting Caddy")
	if err := csc.cmd.Start(); err != nil {
		return err
	}

	var i int
	for ; i < 5; i += 1 {
		secs := int(math.Pow(2, float64(i)))
		csc.logger.Info("waiting for Caddy to be ready...", "seconds", secs)

		time.Sleep(time.Second * time.Duration(secs))

		resp, err := csc.doApiRequest("GET", "/config/", "", nil)

		var status int
		if resp != nil {
			status = resp.StatusCode
		}

		if err == nil && status/100 == 2 {
			break
		}

		csc.logger.Debug("Caddy not ready", "error", err, "status", status)
	}

	if i == 5 {
		return errors.New("failed to start Caddy")
	}

	return nil
}

func (csc *Controller) stop(context.Context) error {
	csc.logger.Info("stopping Caddy")
	if err := csc.cmd.Process.Signal(syscall.SIGINT); err != nil {
		return fmt.Errorf("send interrupt signal to Caddy server: %w", err)
	}
	return csc.cmd.Wait()
}

func (csc *Controller) Reconfigure(routes RouteSpec) error {
	cfg := csc.buildCaddyConfig(routes)

	csc.logger.Debug("applying new Caddy config", "config", cfg)

	resp, err := csc.doApiRequest(http.MethodPost, "/load", "text/caddyfile", cfg)
	if err != nil {
		if errors.Is(err, errFailedRequest) {
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			csc.logger.Error("failed to apply Caddy config", "status", resp.StatusCode, "resp", string(b))
		}
		return fmt.Errorf("apply Caddy config: %w", err)
	}
	defer resp.Body.Close()

	csc.logger.Debug("reconfigured Caddy", "status", resp.StatusCode)
	return nil
}

var errFailedRequest = errors.New("failed request (non-2xx status code)")

func (csc *Controller) doApiRequest(method, path, contentType string, body []byte) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	rq, err := http.NewRequest(method, (&url.URL{
		Scheme: "http",
		Host:   csc.adminApiSocket,
		Path:   path,
	}).String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create HTTP request: %w", err)
	}

	if contentType != "" {
		rq.Header.Set("Content-Type", contentType)
	}
	rq.Close = true

	resp, err := http.DefaultClient.Do(rq)
	if err != nil {
		return nil, fmt.Errorf("do HTTP request: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		return resp, errFailedRequest
	}

	return resp, nil
}

func (csc *Controller) buildCaddyConfig(kr RouteSpec) []byte {
	filesystemCounter := 0
	filesystems := make(map[string]int)

	var rsb bytes.Buffer

	rsb.WriteString(`(canonical_redir) {
	redir @no_trailing_slash {scheme}://{hostport}{http.request.orig_uri.path}/{query_with_mark} permanent
}
`)

	kr.sortValues()
	for domain, routes := range kr {
		rsb.WriteString("http://")
		rsb.WriteString(domain)
		rsb.WriteString(" {\n")

		rsb.WriteString(`@no_trailing_slash {
	not path */
	file {path}/
}
`)

		for _, route := range routes {
			if route.ContentPath == "" {
				continue
			}

			var fsno int
			if v, found := filesystems[route.ContentPath]; found {
				fsno = v
			} else {
				fsno = filesystemCounter
				filesystemCounter += 1
				filesystems[route.ContentPath] = fsno
			}
			fsid := strconv.Itoa(fsno)

			if route.Path != "/" {
				rsb.WriteString("handle_path ")
				rsb.WriteString(strings.TrimSuffix(route.Path, "/"))
				rsb.WriteString("* {\n")
			}

			rsb.WriteString("import canonical_redir\nfile_server {\nfs ")
			rsb.WriteString(fsid)
			rsb.WriteString("\n}\n")

			if route.Path != "/" {
				rsb.WriteString("}\n")
			}
		}

		rsb.WriteString("}\n")
	}

	var gsb bytes.Buffer

	gsb.WriteString(`{
auto_https off
admin `)
	gsb.WriteString(csc.adminApiSocket)
	gsb.WriteString("\nhttp_port ")
	gsb.WriteString(strconv.Itoa(csc.config.HTTP.SitesPort))
	gsb.WriteString("\ndefault_bind ")
	gsb.WriteString(csc.config.HTTP.SitesHost)
	gsb.WriteString("\nlog {\nlevel ")

	if csc.config.Debug {
		gsb.WriteString("DEBUG")
	} else {
		gsb.WriteString("WARN")
	}

	gsb.WriteString("\n}\n")

	for zipfilePath, val := range filesystems {
		gsb.WriteString("filesystem ")
		gsb.WriteString(strconv.Itoa(val))
		gsb.WriteString(" zipfile ")
		gsb.WriteString(fmt.Sprintf("%#v", path.Join(csc.config.Platform.SitesDirectory, zipfilePath)))
		gsb.WriteRune('\n')
	}
	gsb.WriteString("}\n")

	return append(gsb.Bytes(), rsb.Bytes()...)
}
