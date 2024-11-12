package config

import (
	"go.akpain.net/cfger"
)

type HTTP struct {
	ManagementAddress string
	SitesAddress      string
}

type Database struct {
	DSN string
}

type Platform struct {
	SitesDirectory         string
	MaxUploadSizeMegabytes int
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
			ManagementAddress: cl.Get("http.managementAddress").WithDefault("127.0.0.1:8080").AsString(),
			SitesAddress:      cl.Get("http.sitesAddress").WithDefault("127.0.0.1:8081").AsString(),
		},
		Database: &Database{
			DSN: cl.Get("database.dsn").WithDefault("palmatum.db").AsString(),
		},
		Platform: &Platform{
			SitesDirectory:         cl.Get("platform.sitesDirectory").Required().AsString(),
			MaxUploadSizeMegabytes: cl.Get("platform.maxUploadSizeMegabytes").WithDefault(512).AsInt(),
		},
	}

	return conf, nil
}
