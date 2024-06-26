package httpsrv

import (
	"archive/zip"
	_ "embed"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
)

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

	if formFileHeader.Size > 1000*1000*int64(ro.config.Platform.MaxUploadSizeMegabytes) {
		_ = BadRequestResponse(w, fmt.Sprintf("archive too large (maximum size %dMB)", ro.config.Platform.MaxUploadSizeMegabytes))
		return
	}

	zipfileReader, err := zip.NewReader(formFile, formFileHeader.Size)
	if err != nil {
		panic(fmt.Errorf("creating zip file reader: %w", err))
	}

	// We extract the zip file to a temporary location first as to not end up with pages being served with a mix of old
	// and new content while zip file extraction is taking place.

	// Create temporary directory
	tempDir, err := os.MkdirTemp(ro.config.Platform.SitesDirectory, "")
	if err != nil {
		panic(fmt.Errorf("creating temporary directory: %w", err))
	}
	defer os.RemoveAll(tempDir) // by the time this runs we should have copied the directory to its final place but this

	// Extract ZIP file contents
	for _, file := range zipfileReader.File {
		if file.Name[len(file.Name)-1] == '/' {
			continue
		}

		newPath := path.Join(tempDir, file.Name)
		dir := path.Dir(newPath)

		if _, err := os.Stat(dir); err != nil && errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(dir, 0777); err != nil {
				panic(fmt.Errorf("creating directory %s during archive unpacking: %w", dir, err))
			}
		}

		fh, err := os.OpenFile(newPath, os.O_WRONLY|os.O_CREATE, 0777)
		if err != nil {
			panic(fmt.Errorf("creating file %s during archive unpacking: %w", newPath, err))
		}

		zfh, err := file.Open()
		if err != nil {
			panic(fmt.Errorf("opening archive file %s during unpacking: %w", file.Name, err))
		}

		if _, err := io.Copy(fh, zfh); err != nil {
			panic(fmt.Errorf("copying file data to %s during archive unpacking: %w", newPath, err))
		}

		if err := fh.Close(); err != nil {
			panic(fmt.Errorf("closing %s during archive unpacking: %w", newPath, err))
		}

		if err := zfh.Close(); err != nil {
			panic(fmt.Errorf("closing zip file %s during archive unpacking: %w", file.Name, err))
		}
	}

	// Delete old directory (if applicable)
	permPath := path.Join(ro.config.Platform.SitesDirectory, siteName)
	_ = os.RemoveAll(permPath)

	// Move temporary directory to new directory
	if err := os.Rename(tempDir, permPath); err != nil {
		panic(fmt.Errorf("renaming temporary directory: %w", err))
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
