package main

import "net/http"

func (app *Application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", app.homeHandler)
	mux.HandleFunc("GET /motions", app.motionHandler)
	mux.HandleFunc("POST /generate", app.generateHandler)

	return mux
}
