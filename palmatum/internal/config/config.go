package config

import (
	"golang.org/x/exp/slog"
)

type HTTP struct {
	Host string
	Port int
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
	cl := new(configLoader)
	if err := cl.load("config.yml"); err != nil {
		return nil, err
	}

	conf := &Config{
		Debug: asBool(cl.withDefault("debug", false)),
		HTTP: &HTTP{
			Host: asString(cl.withDefault("http.host", "127.0.0.1")),
			Port: asInt(cl.withDefault("http.port", 8080)),
		},
		Database: &Database{
			DSN: asString(cl.withDefault("database.dsn", "website.db")),
		},
		Platform: &Platform{
			SitesDirectory:         asString(cl.required("platform.sitesDirectory")),
			MaxUploadSizeMegabytes: asInt(cl.withDefault("platform.maxUploadSizeMegabytes", 512)),
		},
	}

	if conf.Debug {
		slog.Debug("debug mode enabled")
	}

	return conf, nil
}
