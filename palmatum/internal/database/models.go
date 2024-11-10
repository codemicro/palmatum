package database

type SiteModel struct {
	Slug        string `db:"slug"`
	Name        string `db:"name"`
	ContentPath string `db:"content_path"`
}

type RouteModel struct {
	SiteSlug string `db:"site"`
	Domain   string `db:"domain"`
	Path     string `db:"path"`
}
