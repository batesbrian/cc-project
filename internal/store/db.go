package store

import (
	"database/sql"
	_ "embed"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

func InitSchema(db *sql.DB) error {
	_, err := db.Exec(schemaSQL)
	return err
}

func Open(dsn string) (*sql.DB, error) {
	connString := dsn + "?_pragma=journal_mode(WAL)" +
		"&_pragma=busy_timeout(5000)" + "&_pragma=foreign_keys(ON)"

	db, err := sql.Open("sqlite", connString)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
