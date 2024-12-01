package httpsrv

import (
	"embed"
	_ "embed"
	"errors"
	"fmt"
	"github.com/codemicro/palmatum/palmatum/internal/config"
	"github.com/codemicro/palmatum/palmatum/internal/core"
	"go.uber.org/fx"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

//go:embed static
var staticAssets embed.FS

func NewManagementServer(lc fx.Lifecycle, args ServerArgs) (*http.Server, error) {
	mux := http.NewServeMux()
	mr := managementRoutes{
		logger: args.Logger,
		core:   args.Core,
		config: args.Config,
	}

	lc.Append(fx.Hook{
		OnStart: mr.initManagementTemplates,
	})

	mux.HandleFunc("POST /api/site", handleErrors(args.Logger, mr.apiCreateSite))
	mux.HandleFunc("POST /api/site/bundle", handleErrors(args.Logger, mr.apiUploadSiteBundle))
	mux.HandleFunc("DELETE /api/site", handleErrors(args.Logger, mr.apiDeleteSite))
	mux.HandleFunc("POST /api/site/route", handleErrors(args.Logger, mr.apiCreateRoute))
	mux.HandleFunc("DELETE /api/site/route", handleErrors(args.Logger, mr.apiDeleteRoute))

	mux.HandleFunc("GET /{$}", handleErrors(args.Logger, mr.index))
	mux.HandleFunc("GET /createSite", handleErrors(args.Logger, mr.createSitePartial))
	mux.HandleFunc("GET /uploadSite", handleErrors(args.Logger, mr.uploadSitePartial))
	mux.HandleFunc("GET /deleteSite", handleErrors(args.Logger, mr.deleteSitePartial))
	mux.HandleFunc("GET /addRoute", handleErrors(args.Logger, mr.addRoutePartial))

	{
		subfs, err := fs.Sub(staticAssets, "static")
		if err != nil {
			return nil, fmt.Errorf("subset embedded static asset filesystem: %w", err)
		}
		mux.Handle("GET /", http.FileServer(http.FS(subfs)))
	}

	return newServer(args, args.Config.HTTP.ManagementAddress(), mux), nil
}

type managementRoutes struct {
	logger *slog.Logger
	core   *core.Core
	config *config.Config

	templates *template.Template
}

func (mr *managementRoutes) apiCreateSite(rw http.ResponseWriter, rq *http.Request) error {
	siteSlug := rq.FormValue("slug")
	siteSlug = strings.TrimSpace(siteSlug)

	if len(siteSlug) == 0 {
		_ = badRequestResponse(rw, "Invalid slug (cannot be an empty string)")
		return nil
	}

	_, err := mr.core.CreateSite(siteSlug)
	if err != nil {
		var e *core.Error
		if errors.As(err, &e) {
			_ = badRequestResponse(rw, err.Error())
			return nil
		}
		return fmt.Errorf("create new site: %w", err)
	}

	rw.Header().Set("HX-Refresh", "true")
	rw.WriteHeader(http.StatusCreated)
	return nil
}

func (mr *managementRoutes) apiDeleteSite(rw http.ResponseWriter, rq *http.Request) error {
	siteSlug := rq.FormValue("slug")
	siteSlug = strings.TrimSpace(siteSlug)

	if len(siteSlug) == 0 {
		_ = badRequestResponse(rw, "Invalid slug (cannot be an empty string)")
		return nil
	}

	if err := mr.core.DeleteSite(siteSlug); err != nil {
		return fmt.Errorf("delete site: %w", err)
	}

	rw.Header().Set("HX-Refresh", "true")
	rw.WriteHeader(http.StatusOK)
	return nil
}

func (mr *managementRoutes) apiUploadSiteBundle(rw http.ResponseWriter, rq *http.Request) error {
	siteSlug := strings.TrimSpace(rq.FormValue("slug"))
	if siteSlug == "" {
		_ = badRequestResponse(rw, "Missing slug")
		return nil
	}

	if err := core.ValidateSiteSlug(siteSlug); err != nil {
		_ = badRequestResponse(rw, err.Error())
		return nil
	}

	formFile, formFileHeader, err := rq.FormFile("bundle")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			_ = badRequestResponse(rw, "missing bundle file")
			return nil
		}
		if err.Error() == "request Content-Type isn't multipart/form-data" {
			_ = badRequestResponse(rw, "request Content-Type isn't multipart/form-data")
			return nil
		}
		panic(fmt.Errorf("loading archive request parameter: %w", err))
	}
	defer formFile.Close()

	if formFileHeader.Size > 1000*1000*int64(mr.config.Platform.MaxUploadSizeMegabytes) {
		_ = badRequestResponse(rw, fmt.Sprintf("archive too large (maximum size %dMB)", mr.config.Platform.MaxUploadSizeMegabytes))
		return nil
	}

	contentPath, err := mr.core.IngestSiteArchive(formFile)
	if err != nil {
		return fmt.Errorf("ingest site archive site archive: %w", err)
	}

	if err := mr.core.UpdateContentPath(siteSlug, contentPath); err != nil {
		return fmt.Errorf("update site: %w", err)
	}

	rw.Header().Set("HX-Refresh", "true")
	rw.WriteHeader(http.StatusOK)
	return nil
}

func (mr *managementRoutes) apiCreateRoute(rw http.ResponseWriter, rq *http.Request) error {
	siteSlug := rq.FormValue("slug")
	domain := rq.FormValue("domain")
	path := rq.FormValue("path")

	mr.logger.Debug("create route", "slug", siteSlug, "domain", domain, "path", path)

	_, err := mr.core.CreateRoute(siteSlug, domain, path)
	if err != nil {
		var e *core.Error
		if errors.As(err, &e) {
			_ = badRequestResponse(rw, err.Error())
			return nil
		}
		return fmt.Errorf("create new route: %w", err)
	}

	rw.Header().Set("HX-Refresh", "true")
	rw.WriteHeader(http.StatusCreated)
	return nil
}

func (mr *managementRoutes) apiDeleteRoute(rw http.ResponseWriter, rq *http.Request) error {
	routeIDStr := rq.FormValue("id")
	routeID, err := strconv.Atoi(routeIDStr)
	if err != nil {
		_ = badRequestResponse(rw, "invalid route ID")
		return nil
	}

	if err := mr.core.DeleteRoute(routeID); err != nil {
		return fmt.Errorf("delete route: %w", err)
	}

	rw.Header().Set("HX-Refresh", "true")
	rw.WriteHeader(http.StatusOK)
	return nil
}
