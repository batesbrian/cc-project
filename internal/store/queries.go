package store

import (
	"database/sql"
	"fmt"
	"strings"
)

func GetCaseTypes(db *sql.DB) ([]CaseType, error) {
	rows, err := db.Query(`SELECT id, slug, name FROM case_types`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var caseTypes []CaseType

	for rows.Next() {
		var ct CaseType

		err := rows.Scan(&ct.ID, &ct.Slug, &ct.Name)
		if err != nil {
			return nil, err
		}
		caseTypes = append(caseTypes, ct)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return caseTypes, nil
}

func GetMotion(db *sql.DB, mID int64) (Motion, error) {
	var m Motion
	err := db.QueryRow(`
		SELECT id, slug, name FROM motions WHERE id = ?`,
		mID,
	).Scan(&m.ID, &m.Slug, &m.Name)

	return m, err
}

func GetMotionsByCaseType(db *sql.DB, ctID int64) ([]Motion, error) {
	rows, err := db.Query(
		`SELECT id, slug, name FROM motions
		WHERE case_type_id = ?`,
		ctID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var motions []Motion

	for rows.Next() {
		var m Motion

		err := rows.Scan(&m.ID, &m.Slug, &m.Name)
		if err != nil {
			return nil, err
		}
		motions = append(motions, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return motions, nil
}

func GetIssuesByMotion(db *sql.DB, mID int64) ([]Issue, error) {
	rows, err := db.Query(
		`SELECT id, slug, name, template_path
		FROM issues 
		WHERE motion_id = ? AND active = 1
		ORDER BY sort_order, name`,
		mID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []Issue

	for rows.Next() {
		var i Issue

		err := rows.Scan(&i.ID, &i.Slug, &i.Name, &i.TemplatePath)
		if err != nil {
			return nil, err
		}
		issues = append(issues, i)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return issues, nil
}

func GetIssuesByIDs(db *sql.DB, ids []int64) ([]Issue, error) {
	placeholders := make([]string, len(ids))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(`
		SELECT id, slug, name, template_path
		FROM issues
		WHERE id IN (%s)
		AND active = 1
		ORDER BY sort_order
	`, strings.Join(placeholders, ","))

	rows, err := db.Query(query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []Issue

	for rows.Next() {
		var i Issue

		err := rows.Scan(&i.ID, &i.Slug, &i.Name, &i.TemplatePath)
		if err != nil {
			return nil, err
		}

		issues = append(issues, i)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return issues, nil
}

func GetCaseTypesWithMotions(db *sql.DB) ([]CaseTypeWithMotions, error) {
	rows, err := db.Query(`
		SELECT ct.id, ct.slug, ct.name, m.id, m.slug, m.name
		FROM case_types ct
		JOIN motions m ON m.case_type_id = ct.id
		ORDER BY ct.name, m.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []CaseTypeWithMotions

	for rows.Next() {
		var ct CaseType
		var m Motion

		err := rows.Scan(&ct.ID, &ct.Slug, &ct.Name, &m.ID, &m.Slug, &m.Name)
		if err != nil {
			return nil, err
		}

		n := len(groups)

		if n == 0 || groups[n-1].CaseType.ID != ct.ID {
			groups = append(groups, CaseTypeWithMotions{CaseType: ct})
			n++
		}

		groups[n-1].Motions = append(groups[n-1].Motions, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func GetMotionWithIssues(db *sql.DB, motionID int64) (MotionWithIssues, error) {
	m, err := GetMotion(db, motionID)
	if err != nil {
		return MotionWithIssues{}, err
	}

	issues, err := GetIssuesByMotion(db, motionID)
	if err != nil {
		return MotionWithIssues{}, err
	}

	return MotionWithIssues{Motion: m, Issues: issues}, nil
}

func GetCaseTypeByMotion(db *sql.DB, motionID int64) (string, error) {
	var slug string
	row := db.QueryRow(`
		SELECT ct.slug
		FROM motions m 
		JOIN case_types ct ON ct.id = m.case_type_id
		WHERE m.id = ?
	`, motionID)
	err := row.Scan(&slug)
	if err != nil {
		return "", err
	}
	return slug, nil
}
