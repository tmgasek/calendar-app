package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	app.render(w, http.StatusOK, "home.tmpl", data)
}

// Easily render templates from the cache
func (app *application) render(
	w http.ResponseWriter,
	status int,
	page string,
	data *templateData,
) {
	// retrieve the right template set from cache based on page name
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("The template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	// init new buffer
	buf := new(bytes.Buffer)

	// write template to buffer instead of to the http.ResponseWriter
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// if template written to buffer w/o errors, we are good to go.
	w.WriteHeader(status)

	buf.WriteTo(w)
}

// returns pointer to templateData struct inited with curr year.
func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear: time.Now().Year(),
	}
}

type userLoginForm struct {
	Email    string `form:"email"`
	Password string `form:"password"`
	// validator.Validator `form:"-"`
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, http.StatusOK, "login.tmpl", data)
}
