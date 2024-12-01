package core

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/codemicro/palmatum/palmatum/internal/database"
	"github.com/mattn/go-sqlite3"
	"os"
	"regexp"
	"strings"
)

type Error struct {
	m string
}

func newError(s string) error {
	return &Error{
		m: s,
	}
}

func (e *Error) Error() string {
	return e.m
}

var (
	ErrDuplicateSlug = newError("slug in use")
	ErrInvalidSlug   = newError("invalid slug")

	SiteSlugValidationRegexp = regexp.MustCompile(`^([\w\-._]{2,})$`)
)

func ValidateSiteSlug(s string) error {
	if !SiteSlugValidationRegexp.MatchString(s) {
		return ErrInvalidSlug
	}
	return nil
}

func (c *Core) CreateSite(siteSlug string) (*database.SiteModel, error) {
	if err := ValidateSiteSlug(siteSlug); err != nil {
		return nil, err
	}

	_, err := c.Database.Exec(`INSERT INTO sites(slug) VALUES (?)`, siteSlug)
	if err != nil {
		var e sqlite3.Error
		if errors.As(err, &e) && e.Code == sqlite3.ErrConstraint {
			return nil, ErrDuplicateSlug
		}
		return nil, fmt.Errorf("call database: %w", err)
	}

	return &database.SiteModel{
		Slug: siteSlug,
	}, nil
}

func (c *Core) DeleteSite(siteSlug string) error {
	tx, err := c.Database.Beginx()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM routes WHERE site = ?`, siteSlug); err != nil {
		return fmt.Errorf("delete routes: %w", err)
	}

	var contentPath string

	if err := tx.QueryRow(`DELETE FROM sites WHERE slug = ? RETURNING content_path`, siteSlug).Scan(&contentPath); err != nil {
		return fmt.Errorf("delete sites: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	if contentPath != "" {
		if err := os.Remove(c.getPathOnDisk(contentPath)); err != nil {
			return fmt.Errorf("remove path: %w", err)
		}
	}

	if err := c.BuildKnownRoutes(); err != nil {
		return fmt.Errorf("rebuild known routes: %w", err)
	}

	return nil
}

func (c *Core) UpdateSite(s *database.SiteModel) error {
	_, err := c.Database.Exec(`UPDATE sites SET content_path=? WHERE slug = ?`, s.ContentPath, s.Slug)
	if err != nil {
		return fmt.Errorf("call database: %w", err)
	}

	if err := c.BuildKnownRoutes(); err != nil {
		return fmt.Errorf("rebuild known routes: %w", err)
	}

	return nil
}

var (
	ErrInvalidDomain  = newError("invalid domain")
	ErrInvalidPath    = newError("invalid path (must start with /)")
	ErrRouteNotUnique = newError("route is not unique")
)

func (c *Core) CreateRoute(siteSlug, domain, path string) (*database.RouteModel, error) {
	tx, err := c.Database.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin database transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := database.GetSite(tx, siteSlug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidSlug
		}
		return nil, fmt.Errorf("get site from database: %w", err)
	}

	domain = strings.TrimSpace(domain)

	if domain == "" {
		return nil, ErrInvalidDomain
	}

	if len(path) != 0 {
		if path[0] != '/' {
			return nil, ErrInvalidPath
		}
	} else {
		path = "/"
	}

	var id int

	if err := tx.QueryRowx("INSERT INTO routes(site, domain, path) VALUES (?, ?, ?) RETURNING id", siteSlug, domain, path).Scan(&id); err != nil {
		var e sqlite3.Error
		if errors.As(err, &e) {
			if e.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return nil, ErrInvalidSlug
			}
			if e.ExtendedCode == sqlite3.ErrConstraintUnique {
				return nil, ErrRouteNotUnique
			}
		}
		return nil, fmt.Errorf("call database: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit databse transaction: %w", err)
	}

	if err := c.BuildKnownRoutes(); err != nil {
		return nil, fmt.Errorf("rebuild known routes: %w", err)
	}

	return &database.RouteModel{
		ID:     id,
		Site:   siteSlug,
		Domain: domain,
		Path:   path,
	}, nil
}

func (c *Core) DeleteRoute(id int) error {
	_, err := c.Database.Exec(`DELETE FROM routes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("call database: %w", err)
	}

	if err := c.BuildKnownRoutes(); err != nil {
		return fmt.Errorf("rebuild known routes: %w", err)
	}

	return nil
}
