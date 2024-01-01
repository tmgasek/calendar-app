package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	// "github.com/justinas/alice"
	// "snippetbox.tmgasek.net/ui"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	// create wrapper around our NotFound() helper.
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	router.HandlerFunc(http.MethodGet, "/ping", ping)

	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodGet, "/user/login", app.userLogin)

	return router
}
