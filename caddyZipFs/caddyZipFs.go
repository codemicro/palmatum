package caddyZipFs

import (
	"archive/zip"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"go4.org/readerutil"
	"io"
	"io/fs"
)

func init() {
	caddy.RegisterModule(new(ZipFs))
}

type ZipFs struct {
	SourceZipPath string `json:"path"`
	reader        *zip.ReadCloser
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

func (z *ZipFs) Provision(caddy.Context) (err error) {
	z.reader, err = zip.OpenReader(z.SourceZipPath)
	return
}

func (z *ZipFs) Cleanup() error {
	if z.reader != nil {
		return z.reader.Close()
	}
	return nil
}

type fileWrapper struct {
	io.Seeker
	fs.File
}

func (z *ZipFs) Open(name string) (fs.File, error) {
	f, err := z.reader.Open(name)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat zip file: %w", err)
	}

	// This is required because Caddy secretly needs an io.Seeker from the returned fs.File but doesn't say that
	return &fileWrapper{
		Seeker: readerutil.NewFakeSeeker(f, fi.Size()),
		File:   f,
	}, nil
}

// UnmarshalCaddyfile unmarshals a zipfile instantiation from a Caddyfile.
//
// Example syntax:
//
//	filesystem zf zipfile /path/to/my/website.zip
func (z *ZipFs) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	var arg string
	if !d.Args(&arg) {
		return d.Err("missing ZIP file path")
	}

	z.SourceZipPath = arg

	return nil
}
