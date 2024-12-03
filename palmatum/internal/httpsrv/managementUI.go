package httpsrv

import (
	"context"
	"embed"
	"fmt"
	"github.com/codemicro/palmatum/palmatum/internal/database"
	"html/template"
	"io/fs"
	"net/http"
	"time"
)

//go:embed templates/*
var managementTemplateSource embed.FS

func (mr *managementRoutes) initManagementTemplates(_ context.Context) error {
	mr.templates = template.New("")
	mr.templates.Funcs(map[string]any{
		"fmtTime": func(ti int64) string {
			return time.Unix(ti, 0).Format("2006-01-02 15:04")
		},
	})

	f, err := fs.Sub(fs.FS(managementTemplateSource), "templates")
	if err != nil {
		return fmt.Errorf("subset filesystem: %w", err)
	}

	mr.templates, err = mr.templates.ParseFS(f, "*.html")
	if err != nil {
		return fmt.Errorf("parse templates: %w", err)
	}

	return nil
}

func (mr *managementRoutes) index(rw http.ResponseWriter, rq *http.Request) error {
	var templateData = struct {
		Sites []*database.SiteModel
	}{}

	s, err := database.GetSitesWithRoutes(mr.core.Database)
	if err != nil {
		return fmt.Errorf("get sites with routes: %w", err)
	}
	templateData.Sites = s

	return mr.templates.ExecuteTemplate(rw, "index.html", &templateData)
}

func (mr *managementRoutes) createSitePartial(rw http.ResponseWriter, _ *http.Request) error {
	rw.Header().Set("Hx-Trigger-After-Swap", "showModal")
	return mr.templates.ExecuteTemplate(rw, "createSite.html", nil)
}

func (mr *managementRoutes) uploadSitePartial(rw http.ResponseWriter, rq *http.Request) error {
	rw.Header().Set("Hx-Trigger-After-Swap", "showModal")
	return mr.templates.ExecuteTemplate(rw, "uploadSite.html", rq.URL.Query().Get("slug"))
}

func (mr *managementRoutes) deleteSitePartial(rw http.ResponseWriter, rq *http.Request) error {
	rw.Header().Set("Hx-Trigger-After-Swap", "showModal")
	return mr.templates.ExecuteTemplate(rw, "deleteSite.html", rq.URL.Query().Get("slug"))
}

func (mr *managementRoutes) addRoutePartial(rw http.ResponseWriter, rq *http.Request) error {
	rw.Header().Set("Hx-Trigger-After-Swap", "showModal")
	return mr.templates.ExecuteTemplate(rw, "addRoute.html", rq.URL.Query().Get("slug"))
}

func (mr *managementRoutes) deleteRoutePartial(rw http.ResponseWriter, rq *http.Request) error {
	rw.Header().Set("Hx-Trigger-After-Swap", "showModal")
	return mr.templates.ExecuteTemplate(rw, "deleteRoute.html", struct {
		ID    string
		Route string
	}{ID: rq.URL.Query().Get("id"), Route: rq.URL.Query().Get("domain") + rq.URL.Query().Get("path")})
}
