package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const programSchemaVersion = 1

func New(fname string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", fname)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS schema_version(
		"n" integer not null
	)`)
	if err != nil {
		return nil, fmt.Errorf("create schema_version table: %w", err)
	}

	var currentSchemaVersion int
	if err := db.QueryRowx("SELECT n FROM schema_version").Scan(&currentSchemaVersion); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("unable to read schema version from database: %w", err)
		}
	}

	for currentSchemaVersion < programSchemaVersion {
		switch currentSchemaVersion {
		case 0:
			_, err = db.Exec(`CREATE TABLE sites(
				"slug" varchar not null primary key,
				"content_path" varchar not null
			)`)
			if err != nil {
				return nil, fmt.Errorf("create sites table: %w", err)
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
				return nil, fmt.Errorf("create routes table: %w", err)
			}
			currentSchemaVersion = 1
		case programSchemaVersion:
			// noop
		}
	}

	_, err = db.Exec(`DELETE FROM schema_version`)
	if err != nil {
		return nil, fmt.Errorf("delete old schema version number: %w", err)
	}

	_, err = db.Exec(`INSERT INTO schema_version(n) VALUES (?)`, programSchemaVersion)
	if err != nil {
		return nil, fmt.Errorf("insert schema version number: %w", err)
	}

	return db, nil
}
