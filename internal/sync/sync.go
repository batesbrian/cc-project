package sync

import (
	"database/sql"
	"fmt"
	"io/fs"
	"strings"
	"time"
)

type iPath struct {
	ct   string
	m    string
	g    string
	i    string
	path string
}

func parseIssuePath(path string) (iPath, bool) {
	if !strings.HasSuffix(path, ".docx") {
		return iPath{}, false
	}

	parts := strings.Split(path, "/")
	if len(parts) != 4 {
		return iPath{}, false
	}

	return iPath{
		ct:   parts[0],
		m:    parts[1],
		g:    parts[2],
		i:    strings.TrimSuffix(parts[3], ".docx"),
		path: path,
	}, true
}

func SyncTemplates(db *sql.DB, fsys fs.FS) error {
	syncToken := time.Now().UTC().Format(time.RFC3339Nano)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start sync tx: %w", err)
	}
	defer tx.Rollback()

	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ip, ok := parseIssuePath(path)
		if !ok {
			return nil
		}

		return upsertIssue(tx, ip, syncToken)
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
