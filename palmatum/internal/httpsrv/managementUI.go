package httpsrv

import (
	"context"
	"embed"
	"fmt"
	"github.com/codemicro/palmatum/palmatum/internal/database"
	"io/fs"
	"net/http"
)

//go:embed templates/*
var managementTemplateSource embed.FS

func (mr *managementRoutes) initManagementTemplates(_ context.Context) error {
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
	return mr.templates.ExecuteTemplate(rw, "createSite.html", nil)
}

func (mr *managementRoutes) uploadSitePartial(rw http.ResponseWriter, rq *http.Request) error {
	return mr.templates.ExecuteTemplate(rw, "uploadSite.html", rq.URL.Query().Get("slug"))
}

func (mr *managementRoutes) addRoutePartial(rw http.ResponseWriter, rq *http.Request) error {
	return mr.templates.ExecuteTemplate(rw, "addRoute.html", rq.URL.Query().Get("slug"))
}
