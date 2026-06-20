package sync

import (
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

func SyncTemplates(db *sql.DB, root string) error {
	syncToken := time.Now().UTC().Format(time.RFC3339Nano)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin sync tx: %w", err)
	}
	defer tx.Rollback()

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".docx" {
			return err
		}

		rel, _ := filepath.Rel(root, path)
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) != 3 {
			return nil
		}

		ctSlug := parts[0]
		mSlug := parts[1]
		iSlug := strings.TrimSuffix(parts[2], ".docx")

		return upsertIssue(tx, ctSlug, mSlug, iSlug, rel, syncToken)
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
