package app

import (
	"database/sql"
	"log/slog"

	"github.com/batesbrian/cc-templates/internal/docx"
)

type Application struct {
	Logger *slog.Logger
	Store  *sql.DB
	Gen    docx.Generator
}

func NewApplication(logger *slog.Logger, db *sql.DB, gen docx.Generator) (*Application, error) {
	return &Application{
		Logger: logger,
		Store:  db,
		Gen:    gen,
	}, nil
}
