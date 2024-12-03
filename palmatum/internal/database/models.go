package database

import (
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

type SiteModel struct {
	Slug          string `db:"slug"` // primary key
	ContentPath   string `db:"content_path"`
	LastUpdatedAt int64  `db:"last_updated_at"`

	Routes []*RouteModel `db:"-"`
}

func GetSite(db sqlx.Queryer, slug string) (*SiteModel, error) {
	res := new(SiteModel)
	if err := db.QueryRowx(`SELECT * from sites WHERE "slug" = ?`, slug).StructScan(res); err != nil {
		return nil, err
	}
	return res, nil
}

func GetSites(db sqlx.Queryer) ([]*SiteModel, error) {
	var res []*SiteModel
	if err := sqlx.Select(db, &res, "SELECT * FROM sites"); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return res, nil
}

func GetSitesWithRoutes(db sqlx.Queryer) ([]*SiteModel, error) {
	sites, err := GetSites(db)
	if err != nil {
		return nil, err
	}

	smap := make(map[string]*SiteModel)
	for _, v := range sites {
		smap[v.Slug] = v
	}

	var routes []*RouteModel
	if err := sqlx.Select(db, &routes, "SELECT id, site, domain, path FROM routes"); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	for _, r := range routes {
		smap[r.Site].Routes = append(smap[r.Site].Routes, r)
	}

	return sites, nil
}

type RouteModel struct {
	ID     int    `db:"id"`
	Site   string `db:"site"`
	Domain string `db:"domain"`
	Path   string `db:"path"`
}
