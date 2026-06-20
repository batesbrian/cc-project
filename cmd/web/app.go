package main

import (
	"database/sql"
	"log/slog"
)

type Application struct {
	Logger       *slog.Logger
	Store        *sql.DB
	TemplateRoot string
	// TODO: HTML?
}

func NewApplication(logger *slog.Logger, db *sql.DB, templateRoot string) (*Application, error) {
	return &Application{
		Logger:       logger,
		Store:        db,
		TemplateRoot: templateRoot,
	}, nil
}
