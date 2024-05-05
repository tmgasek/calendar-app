package main

import "net/http"

func (app *application) searchUsers(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")

	users, err := app.models.Users.SearchUsers(query)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Users = users

	app.render(w, http.StatusOK, "users.tmpl", data)
}
