package httpsrv

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/codemicro/palmatum/palmatum/internal/config"
	"github.com/codemicro/palmatum/palmatum/internal/core"
	"github.com/codemicro/palmatum/palmatum/internal/database"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"net/http"
	"os"
	"path"
	"regexp"
)

func NewManagementServer(args ServerArgs) *http.Server {
	return newServer(args, fmt.Sprintf("%s:%d", args.Config.HTTP.Host, args.Config.HTTP.Port), New(args.Config, args.Core))
}

func New(conf *config.Config, c *core.Core) http.Handler {
	r := &routes{
		config: conf,
		core:   c,
	}

	router := httprouter.New()

	router.GET("/-/", r.managementIndex)
	router.POST("/-/upload", r.uploadSite)
	router.POST("/-/delete", r.deleteSite)

	return router
}

type routes struct {
	config *config.Config
	core   *core.Core
}

//go:embed managementIndex.html
var managementPageTemplateSource string
var managementPageTemplate *template.Template

func init() {
	managementPageTemplate = template.New("management page")
	template.Must(managementPageTemplate.Parse(managementPageTemplateSource))
}

func (ro *routes) managementIndex(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {

	dirEntries, err := os.ReadDir(ro.config.Platform.SitesDirectory)
	if err != nil {
		panic(fmt.Errorf("reading sites directory contents: %w", err))
	}

	var sites []string
	for _, de := range dirEntries {
		sites = append(sites, de.Name())
	}

	var templateArgs = struct {
		ActiveSites []string
	}{
		ActiveSites: sites,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := managementPageTemplate.Execute(w, templateArgs); err != nil {
		panic(fmt.Errorf("rendering management page: %w", err))
	}
}

var siteNameValidationRegexp = regexp.MustCompile(`^([\w\-.~!$&'()*+,;=:@]{2,})|(-[\w\-.~!$&'()*+,;=:@]+)$`)

func (ro *routes) uploadSite(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	siteName := r.FormValue("siteName")
	if siteName == "" {
		_ = BadRequestResponse(w, "missing site name")
		return
	}

	if !siteNameValidationRegexp.MatchString(siteName) {
		_ = BadRequestResponse(w, "invalid site name")
		return
	}

	formFile, formFileHeader, err := r.FormFile("archive")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			_ = BadRequestResponse(w, "missing archive file")
			return
		}
		if err.Error() == "request Content-Type isn't multipart/form-data" {
			_ = BadRequestResponse(w, "request Content-Type isn't multipart/form-data")
			return
		}
		panic(fmt.Errorf("loading archive request parameter: %w", err))
	}
	defer formFile.Close()

	if formFileHeader.Size > 1000*1000*int64(ro.config.Platform.MaxUploadSizeMegabytes) {
		_ = BadRequestResponse(w, fmt.Sprintf("archive too large (maximum size %dMB)", ro.config.Platform.MaxUploadSizeMegabytes))
		return
	}

	contentPath, err := ro.core.IngestSiteArchive(formFile)
	if err != nil {
		panic(fmt.Errorf("ingesting site archive: %w", err))
	}

	if err := ro.core.UpsertSite(&database.SiteModel{
		Slug:        siteName,
		ContentPath: contentPath,
	}); err != nil {
		panic(fmt.Errorf("updating site in database: %w", err))
	}

	if IsBrowser(r) {
		w.Header().Add("Location", "/-/")
		w.WriteHeader(303) // 303 See Other - resets the method to GET
	} else {
		w.WriteHeader(204)
	}
}

func (ro *routes) deleteSite(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	siteName := r.FormValue("siteName")
	if !siteNameValidationRegexp.MatchString(siteName) {
		_ = BadRequestResponse(w, "invalid site name")
		return
	}
	_ = os.RemoveAll(path.Join(ro.config.Platform.SitesDirectory, siteName))
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(204)
}
