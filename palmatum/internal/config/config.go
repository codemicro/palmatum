package config

import (
	"fmt"
	"go.akpain.net/cfger"
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
			CaddyExecutablePath:    cl.Get("platform.caddyExecutablePath").WithDefault("caddy").AsString(),
		},
	}

	return conf, nil
}
