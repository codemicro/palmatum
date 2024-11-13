package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/codemicro/palmatum/palmatum/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/fx"
)

const programSchemaVersion = 1

func New(lc fx.Lifecycle, conf *config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", conf.Database.DSN)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			_, err = db.Exec(`CREATE TABLE IF NOT EXISTS schema_version(
				"n" integer not null
			)`)
			if err != nil {
				return fmt.Errorf("create schema_version table: %w", err)
			}

			var currentSchemaVersion int
			if err := db.QueryRowx("SELECT n FROM schema_version").Scan(&currentSchemaVersion); err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					return fmt.Errorf("unable to read schema version from database: %w", err)
				}
			}

			for currentSchemaVersion < programSchemaVersion {
				switch currentSchemaVersion {
				case 0:
					_, err = db.Exec(`CREATE TABLE sites(
						"slug" varchar primary key,
						"content_path" varchar default ''
					)`)
					if err != nil {
						return fmt.Errorf("create sites table: %w", err)
					}

					_, err = db.Exec(`CREATE TABLE routes(
						"id" integer primary key autoincrement,
						"site" varchar not null,
						"domain" varchar not null,
						"path" varchar default '',
						
						foreign key (site) references sites(slug),
						unique (domain, path)
					)`)
					if err != nil {
						return fmt.Errorf("create routes table: %w", err)
					}
					currentSchemaVersion = 1
				case programSchemaVersion:
					// noop
				}
			}

			_, err = db.Exec(`DELETE FROM schema_version`)
			if err != nil {
				return fmt.Errorf("delete old schema version number: %w", err)
			}

			_, err = db.Exec(`INSERT INTO schema_version(n) VALUES (?)`, programSchemaVersion)
			if err != nil {
				return fmt.Errorf("insert schema version number: %w", err)
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return db.Close()
		},
	})

	return db, nil
}
