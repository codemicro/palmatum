package config

import (
	"fmt"
	"go.akpain.net/cfger"
	"os"
	"path"
)

type HTTP struct {
	ManagementPort int
	ManagementHost string
	SitesPort      int
	SitesHost      string
}

func (h *HTTP) ManagementAddress() string {
	return fmt.Sprintf("%s:%d", h.ManagementHost, h.ManagementPort)
}

func (h *HTTP) SitesAddress() string {
	return fmt.Sprintf("%s:%d", h.SitesHost, h.SitesPort)
}

type Database struct {
	DSN string
}

type Platform struct {
	SitesDirectory         string
	MaxUploadSizeMegabytes int
	CaddyExecutablePath    string
}

type Config struct {
	Debug    bool
	HTTP     *HTTP
	Database *Database
	Platform *Platform
}

func Load() (*Config, error) {
	exePath, err := os.Executable()
	if err != nil {
		// If we can't get our own executable location, then we effectively drop back to looking in the current working
		// directory by setting this to an empty SIM.
		//
		// I did check and the empty string interacts nicely with path.Join and won't add random extra copies of /
		//
		// We're adding the empty string just to be safe :)
		_, _ = fmt.Fprintf(os.Stderr, "Warning: unable to determine own executable path: %v", err)
		exePath = ""
	}

	cl := new(cfger.ConfigLoader)
	if err := cl.Load("config.yml"); err != nil {
		return nil, err
	}

	conf := &Config{
		Debug: cl.Get("debug").WithDefault(false).AsBool(),
		HTTP: &HTTP{
			ManagementHost: cl.Get("http.managementHost").WithDefault("127.0.0.1").AsString(),
			ManagementPort: cl.Get("http.managementPort").WithDefault(8080).AsInt(),
			SitesHost:      cl.Get("http.sitesHost").WithDefault("0.0.0.0").AsString(),
			SitesPort:      cl.Get("http.sitesPort").WithDefault(8081).AsInt(),
		},
		Database: &Database{
			DSN: cl.Get("database.dsn").WithDefault("palmatum.db").AsString(),
		},
		Platform: &Platform{
			SitesDirectory:         cl.Get("platform.sitesDirectory").Required().AsString(),
			MaxUploadSizeMegabytes: cl.Get("platform.maxUploadSizeMegabytes").WithDefault(512).AsInt(),
			CaddyExecutablePath:    cl.Get("platform.caddyExecutablePath").WithDefault(path.Join(exePath, "caddy")).AsString(),
		},
	}

	return conf, nil
}
