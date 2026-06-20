package sync

import (
	"database/sql"
	"fmt"
	"strings"
)

func upsertIssue(tx *sql.Tx, ctSlug, mSlug, iSlug, tPath, syncToken string) error {
	ctID, err := upsertCaseType(tx, ctSlug)
	if err != nil {
		return fmt.Errorf("upsert case type %q: %w", ctSlug, err)
	}

	mID, err := upsertMotion(tx, ctID, mSlug)
	if err != nil {
		return fmt.Errorf("upsert motion %q: %w", mSlug, err)
	}

	err = upsertIssueRow(tx, mID, iSlug, tPath, syncToken)
	if err != nil {
		return fmt.Errorf("upsert issue row %q: %w", iSlug, err)
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
	if err != sql.ErrNoRows {
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
	if err != sql.ErrNoRows {
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

func upsertIssueRow(tx *sql.Tx, mID int64, slug, tPath, syncToken string) error {
	var id int64
	err := tx.QueryRow(
		`SELECT id FROM issues
		WHERE motion_id = ? AND slug = ?`,
		mID, slug,
	).Scan(&id)
	if err == sql.ErrNoRows {
		_, err := tx.Exec(
			`INSERT INTO issues
			(motion_id, slug, name, template_path, last_seen)
			VALUES (?, ?, ?, ?, ?)`,
			mID, slug, formatName(slug), tPath, syncToken,
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
