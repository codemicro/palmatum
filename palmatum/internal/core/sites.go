package core

import (
	"errors"
	"fmt"
	"github.com/codemicro/palmatum/palmatum/internal/database"
	"github.com/mattn/go-sqlite3"
)

func (c *Core) UpsertSite(sm *database.SiteModel) error {
	tx, err := c.Database.Beginx()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := database.InsertSite(tx, sm); err != nil {
		if errors.Is(err, sqlite3.ErrConstraint) {
			if err = database.UpdateSite(tx, sm); err != nil {
				return fmt.Errorf("update site: %w", err)
			}
		} else {
			return fmt.Errorf("insert site: %w", err)
		}
	}

	return nil
}
