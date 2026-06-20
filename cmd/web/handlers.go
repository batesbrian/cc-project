package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/batesbrian/cc-templates/internal/store"
)

func (app *Application) homeHandler(w http.ResponseWriter, r *http.Request) {
	groups, err := store.GetCaseTypesWithMotions(app.Store)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// TODO: render home page
}

func (app *Application) motionHandler(w http.ResponseWriter, r *http.Request) {
	motionID := r.URL.Query().Get("motion_id")
	id, err := strconv.ParseInt(motionID, 10, 64)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	motionWithIssues, err := store.GetMotionWithIssues(app.Store, id)
	if err == sql.ErrNoRows {
		app.notFound(w, r, err)
		return
	}

	// TODO: render motion form
}

func (app *Application) notFound(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Error("http not found",
		"method", r.Method,
		"uri", r.URL.RequestURI(),
		"err", err)

	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func (app *Application) badRequest(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Error("http bad request",
		"method", r.Method,
		"uri", r.URL.RequestURI(),
		"err", err)

	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func (app *Application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Error("server error",
		"method", r.Method,
		"uri", r.URL.RequestURI(),
		"err", err)

	http.Error(w, http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError)
}
