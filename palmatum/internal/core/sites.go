package core

import (
	"errors"
	"fmt"
	"github.com/codemicro/palmatum/palmatum/internal/database"
	"github.com/mattn/go-sqlite3"
	"os"
)

func (c *Core) UpsertSite(sm *database.SiteModel) error {

	// TODO: trash all of this :(

	// NOTE TO FUTURE SELF: LOCK BEFORE YOU BEGIN A TRANSACTION. :)

	tx, err := c.Database.Beginx()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := database.InsertSite(tx, sm); err != nil {
		var e sqlite3.Error
		if errors.As(err, &e) && e.Code == sqlite3.ErrConstraint {
			goto exists
		}
		return fmt.Errorf("insert site: %w", err)
	}

	// TODO: rebuild routing graph here

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil

exists:
	existingSite, err := database.GetSite(tx, sm.Slug)
	if err != nil {
		return fmt.Errorf("get existing site: %w", err)
	}

	if err := database.UpdateSite(tx, sm); err != nil {
		return fmt.Errorf("update site: %w", err)
	}

	// TODO: rebuild routing graph here

	if sm.ContentPath != existingSite.ContentPath {
		if err := os.Remove(c.getPathOnDisk(existingSite.ContentPath)); err != nil {
			return fmt.Errorf("remove old site content path: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
