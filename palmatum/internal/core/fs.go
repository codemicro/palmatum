package core

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"os"
)

func (c *Core) IngestSiteArchive(archive io.Reader) (string, error) {
	var (
		key             uuid.UUID
		fname           string
		destinationFile *os.File
	)

	for destinationFile == nil { // take note of the os.O_EXCL flag here - if the file already exists, os.OpenFile will
		// error and cause this loop to be re-run, effectively working as an atomic
		// check-and-create-if-not-exists-else-try-a-different-name step
		key = uuid.New()
		fname = fmt.Sprintf("%s.zip", key)

		f, err := os.OpenFile(c.getPathOnDisk(fname), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			if errors.Is(err, os.ErrExist) {
				continue
			}
			return "", fmt.Errorf("open destination file: %w", err)
		}
		destinationFile = f
	}
	defer destinationFile.Close()

	if _, err := io.Copy(destinationFile, archive); err != nil {
		return "", fmt.Errorf("copy archive to destination file %s: %w", fname, err)
	}

	return fname, nil
}
