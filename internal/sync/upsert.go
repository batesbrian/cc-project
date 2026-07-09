package sync

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func upsertIssue(tx *sql.Tx, ip iPath, syncToken string) error {
	ctID, err := upsertCaseType(tx, ip.ct)
	if err != nil {
		return fmt.Errorf("upsert case type %q: %w", ip.ct, err)
	}

	mID, err := upsertMotion(tx, ctID, ip.m)
	if err != nil {
		return fmt.Errorf("upsert motion %q: %w", ip.m, err)
	}

	gID, err := upsertGroup(tx, mID, ip.g)
	if err != nil {
		return fmt.Errorf("upsert group %q: %w", ip.g, err)
	}

	err = upsertIssueRow(tx, gID, ip.i, ip.path, syncToken)
	if err != nil {
		return fmt.Errorf("upsert issue row %q: %w", ip.i, err)
	}

	return err
}

func upsertCaseType(tx *sql.Tx, slug string) (int64, error) {
	var id int64
	err := tx.QueryRow(
		`SELECT id FROM case_types
		WHERE slug = ?`,
		slug,
	).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	res, err := tx.Exec(
		`INSERT INTO case_types
		(slug, name) VALUES (?, ?)`,
		slug, formatName(slug),
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func upsertMotion(tx *sql.Tx, ctID int64, mSlug string) (int64, error) {
	var id int64
	err := tx.QueryRow(
		`SELECT id FROM motions
		WHERE case_type_id = ? AND slug = ?`,
		ctID, mSlug,
	).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	res, err := tx.Exec(
		`INSERT INTO motions
		(case_type_id, slug, name)
		VALUES (?, ?, ?)`,
		ctID, mSlug, formatName(mSlug),
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func upsertGroup(tx *sql.Tx, mID int64, gSlug string) (int64, error) {
	var id int64
	err := tx.QueryRow(
		`SELECT id FROM groups
		WHERE motion_id = ? AND slug = ?`,
		mID, gSlug,
	).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	res, err := tx.Exec(
		`INSERT INTO groups
		(motion_id, slug, name)
		VALUES (?, ?, ?)`,
		mID, gSlug, formatName(gSlug),
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func upsertIssueRow(tx *sql.Tx, gID int64, slug, tPath, syncToken string) error {
	var id int64
	err := tx.QueryRow(
		`SELECT id FROM issues
		WHERE group_id = ? AND slug = ?`,
		gID, slug,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		_, err := tx.Exec(
			`INSERT INTO issues
			(group_id, slug, name, template_path, last_seen)
			VALUES (?, ?, ?, ?, ?)`,
			gID, slug, formatName(slug), tPath, syncToken,
		)

		return err
	}

	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`UPDATE issues
		SET template_path = ?, last_seen = ? WHERE id = ?`,
		tPath, syncToken, id)

	return err
}

func markOrphans(tx *sql.Tx, syncToken string) error {
	_, err := tx.Exec(
		`UPDATE issues
		SET active = CASE WHEN last_seen = ? THEN 1 ELSE 0 END`,
		syncToken,
	)
	return err
}

func formatName(s string) string {
	words := strings.Split(s, "_")
	for i, w := range words {
		if w == "" {
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}

	return strings.Join(words, " ")
}
