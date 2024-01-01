package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	// "github.com/justinas/alice"
	// "snippetbox.tmgasek.net/ui"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	return router
}
