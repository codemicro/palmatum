package caddyZipFs

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"io/fs"
	"os"
	"path"
	"strings"
	"sync"
)

func init() {
	caddy.RegisterModule(new(ZipFs))
}

type ZipFs struct {
	fileCacheLock sync.Mutex
	fileCache     map[string]*zip.ReadCloser
}

var (
	_ fs.FS              = (*ZipFs)(nil)
	_ caddy.Provisioner  = (*ZipFs)(nil)
	_ caddy.CleanerUpper = (*ZipFs)(nil)
)

func (*ZipFs) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "caddy.fs.zipfile",
		New: func() caddy.Module { return new(ZipFs) },
	}
}

func (z *ZipFs) Provision(caddy.Context) error {
	z.fileCache = make(map[string]*zip.ReadCloser)
	return nil
}

func (z *ZipFs) Cleanup() error {
	// No locking is done here since Caddy calls this after the module is no longer in use, ie. no concurrent access
	for _, v := range z.fileCache {
		if err := v.Close(); err != nil {
			return fmt.Errorf("close cached ZIP file: %w", err)
		}
	}
	return nil
}

func (z *ZipFs) Open(combinedFilePath string) (fs.File, error) {
	zr, fname, err := z.getZipReaderAndInternalFilePath(combinedFilePath)
	if err != nil {
		return nil, err
	}
	return zr.Open(fname)
}

func (z *ZipFs) getZipReaderAndInternalFilePath(combinedFilePath string) (*zip.ReadCloser, string, error) {
	z.fileCacheLock.Lock()
	defer z.fileCacheLock.Unlock()

	for prefix, zipReader := range z.fileCache {
		if strings.HasPrefix(combinedFilePath, prefix) {
			return zipReader, combinedFilePath[len(prefix):], nil
		}
	}

	zipFilePath, internalFilePath, err := splitCombinedZipFilePath(combinedFilePath)
	if err != nil {
		return nil, "", fmt.Errorf("split combined file path: %w", err)
	}

	zr, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return nil, "", fmt.Errorf("open zip file: %w", err)
	}

	z.fileCache[zipFilePath] = zr

	return zr, internalFilePath, nil
}

// splitCombinedZipFilePath splits a combination of a path to a ZIP file on the host OS and a filepath within that ZIP
// file into its constituent parts.
//
// For example, it will turn /srv/websites/mysite.zip/assets/css/main.css into /srv/website/mysite.zip and
// /assets/css/main.css
//
// The first segment of the combined file path that ends in .zip or .ZIP is checked to see if it is a file on the host
// filesystem. If so, the path is split there - otherwise, the path is checked, item by item, until the first
// non-directory entry is found, at which point it is assumed to be the ZIP file and the path is split there.
func splitCombinedZipFilePath(combinedFilePath string) (zipFilePath string, internalFileName string, err error) {
	// find .zip file extension in the path
	// i runs to the length minus for to allow for the length of the file extension
	filePathComponents := strings.Split(combinedFilePath, "/")

	zipFileNameBoundary := -1

	for i, component := range filePathComponents {
		if strings.HasSuffix(component, ".zip") || strings.HasSuffix(component, ".ZIP") {
			zipFileNameBoundary = i + 1
			break
		}
	}

	for i, n := -1, 0; i < len(filePathComponents); i += 1 {
		if i == -1 {
			if zipFileNameBoundary == -1 {
				continue
			}
			n = zipFileNameBoundary
		} else if i == zipFileNameBoundary {
			continue
		} else {
			n = i
		}

		candidateZipFile := path.Join(filePathComponents[:n]...)
		if isFile, err := isPathAFile(candidateZipFile); err != nil {
			return "", "", fmt.Errorf("check candidate ZIP file: %w", err)
		} else if isFile {
			return candidateZipFile, path.Join(filePathComponents[n:]...), nil
		}
	}

	return "", "", errors.New("no ZIP file located")
}

func isPathAFile(fpath string) (bool, error) {
	if _, err := os.Stat(fpath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
