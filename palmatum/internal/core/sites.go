package core

import (
	"errors"
	"fmt"
	"github.com/codemicro/palmatum/palmatum/internal/database"
	"github.com/mattn/go-sqlite3"
	"regexp"
)

var (
	ErrDuplicateSlug = errors.New("slug in use")
	ErrInvalidSlug   = errors.New("invalid slug")

	SiteSlugValidationRegexp = regexp.MustCompile(`^([\w\-.~!$&'()*+,;=:@]{2,})$`)
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
	_, err := c.Database.Exec(`DELETE FROM sites WHERE slug = ?`, siteSlug)
	if err != nil {
		return fmt.Errorf("call database: %w", err)
	}
	return nil
}

func (c *Core) UpdateSite(s *database.SiteModel) error {
	_, err := c.Database.Exec(`UPDATE sites SET content_path=? WHERE slug = ?`, s.ContentPath, s.Slug)
	if err != nil {
		return fmt.Errorf("call database: %w", err)
	}
	return nil
}
