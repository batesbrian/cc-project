package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/batesbrian/cc-templates/internal/docx"
	"github.com/batesbrian/cc-templates/internal/store"
)

func (app *Application) homeHandler(w http.ResponseWriter, r *http.Request) {
	groups, err := store.GetCaseTypesWithMotions(app.Store)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// TODO: render home page
	fmt.Printf("%v\n", groups)
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
	fmt.Printf("%v\n", motionWithIssues)
}

func (app *Application) generateHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	motionString := r.FormValue("motion_id")
	motionInt, err := strconv.ParseInt(motionString, 10, 64)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	ctSlug, err := store.GetCaseTypeByMotion(app.Store, motionInt)
	if err == sql.ErrNoRows {
		app.notFound(w, r, err)
		return
	}
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	caption, ok := docx.GetCaption(ctSlug)
	if !ok {
		app.badRequest(w, r, fmt.Errorf("no caption for case type: %s\n", ctSlug))
		return
	}

	county := r.FormValue("county")
	caption.County = county

	stringIDs := r.Form["issue_ids"]
	if len(stringIDs) == 0 {
		app.badRequest(w, r, errors.New("no issues selected"))
		return
	}
	var intIDs []int64

	for i := range stringIDs {
		id, err := strconv.ParseInt(stringIDs[i], 10, 64)
		if err != nil {
			app.badRequest(w, r, err)
			return
		}

		intIDs = append(intIDs, id)
	}

	issues, err := store.GetIssuesByIDs(app.Store, intIDs)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	var paths []string
	for _, iss := range issues {
		paths = append(paths, iss.TemplatePath)
	}

	changeFont := r.FormValue("font") == "Bookman Old Style"
	changeCitations := r.FormValue("citations") == "underline"

	doc := docx.Docx{
		Caption:         caption,
		Issues:          paths,
		ChangeFont:      changeFont,
		ChangeCitations: changeCitations,
	}

	var buf bytes.Buffer

	err = app.Gen.Build(&buf, doc)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=output.docx")
	w.Header().Set("Content-Type",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document")

	io.Copy(w, &buf)
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
