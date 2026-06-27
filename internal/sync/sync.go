package sync

import (
	"database/sql"
	"fmt"
	"io/fs"
	"strings"
	"time"
)

func SyncTemplates(db *sql.DB, fsys fs.FS) error {
	syncToken := time.Now().UTC().Format(time.RFC3339Nano)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start sync tx: %w", err)
	}
	defer tx.Rollback()

	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || strings.HasSuffix(path, ".docx") {
			return err
		}

		parts := strings.Split(path, "/")
		if len(parts) != 3 {
			return nil
		}

		ctSlug := parts[0]
		mSlug := parts[1]
		iSlug := strings.TrimSuffix(parts[2], ".docx")

		return upsertIssue(tx, ctSlug, mSlug, iSlug, path, syncToken)
	})
	if err != nil {
		return err
	}

	err = markOrphans(tx, syncToken)
	if err != nil {
		return err
	}

	return tx.Commit()
}
