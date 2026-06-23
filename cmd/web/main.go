package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/batesbrian/cc-templates/internal/store"
	"github.com/batesbrian/cc-templates/internal/sync"
)

func main() {
	addr := flag.String("addr", ":8080", "http listening port")
	dsn := flag.String("dsn", "app.db", "SQLite data source name")
	templateRoot := flag.String("templates", "./templates", "docx template directory")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	db, err := store.Open(*dsn)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		panic(err)
	}
	defer db.Close()

	store.InitSchema(db)

	err = sync.SyncTemplates(db, *templateRoot)
	if err != nil {
		logger.Error("template sync failed", "error", err)
		panic(err)
	}

	_, err = NewApplication(logger, db, *templateRoot)
	if err != nil {
		logger.Error("failed to start app", "error", err)
		panic(err)
	}

	logger.Info("starting server", "addr", *addr)

	// err = http.ListenAndServe(*addr, app.routes())
	// logger.Error("server stopped", "err", err)
}
