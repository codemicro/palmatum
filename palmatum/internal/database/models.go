package database

import (
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

type SiteModel struct {
	Slug        string `db:"slug"` // primary key
	ContentPath string `db:"content_path"`

	Routes []*RouteModel `db:"-"`
}

func GetSite(db sqlx.Queryer, slug string) (*SiteModel, error) {
	res := new(SiteModel)
	if err := db.QueryRowx(`SELECT "slug", "content_path" from sites WHERE "slug" = ?`, slug).Scan(res); err != nil {
		return nil, err
	}
	return res, nil
}

func InsertSite(db sqlx.Ext, site *SiteModel) error {
	_, err := sqlx.NamedExec(db, `INSERT INTO sites("slug", "content_path") VALUES(:slug, :content_path)`, site)
	return err
}

func UpdateSite(db sqlx.Ext, site *SiteModel) error {
	_, err := sqlx.NamedExec(db, `UPDATE sites SET content_path = :content_path WHERE slug = :slug`, site)
	return err
}

func GetSites(db sqlx.Queryer) ([]*SiteModel, error) {
	var res []*SiteModel
	if err := sqlx.Select(db, &res, "SELECT slug, content_path FROM sites"); err != nil && !errors.Is(err, sql.ErrNoRows) {
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
	if err := sqlx.Select(db, &routes, "SELECT site, domain, path FROM routes"); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	for _, r := range routes {
		smap[r.Site].Routes = append(smap[r.Site].Routes, r)
	}

	return sites, nil
}

type RouteModel struct {
	Site   string `db:"site"`
	Domain string `db:"domain"`
	Path   string `db:"path"`
}
