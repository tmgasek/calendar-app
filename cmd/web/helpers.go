package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
)

// returns pointer to templateData struct inited with curr year.
func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		IsAuthenticated: app.isAuthenticated(r),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		CSRFToken:       nosurf.Token(r),
		UserId:          app.sessionManager.GetInt(r.Context(), "authenticatedUserID"),
	}
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

	_, err = buf.WriteTo(w)
	if err != nil {
		app.serverError(w, err)
	}
}

// write error msg and stack trace to errorLog, send generic 500 res to user
func (app *application) serverError(w http.ResponseWriter, err error) {
	fmt.Println(err)
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	http.Error(
		w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError,
	)
}

// send specific status code and corresponding description to user
// func (app *application) clientError(w http.ResponseWriter, status int) {
// 	http.Error(w, http.StatusText(status), status)
// }

type ErrorData struct {
	Status  int
	Message string
}

func (app *application) clientError(w http.ResponseWriter, status int, message string) {
	data := &ErrorData{
		Status:  status,
		Message: message,
	}
	app.render(w, status, "error.tmpl", &templateData{ErrorData: data})
}

// convenient wrapper around clientError, sends 404 res to user
// func (app *application) notFound(w http.ResponseWriter) {
// 	app.clientError(w, http.StatusNotFound)
// }

// DST is target destination that we want to decode the form data into.
func (app *application) decodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		// we want to panic if dst invalid
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		// return all other errs as normal
		return err
	}

	return nil
}

// Return true if curr req is coming from an authenticated user, else false.
func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}

	return isAuthenticated
}
