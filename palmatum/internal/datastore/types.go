package datastore

import "time"

type SiteData struct {
	Name string
	LastModified time.Time
	LastModifiedBy string
}