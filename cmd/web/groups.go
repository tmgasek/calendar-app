package main

import (
	"net/http"

	"github.com/tmgasek/calendar-app/internal/validator"
)

func (app *application) viewGroupsPage(w http.ResponseWriter, r *http.Request) {
	templateData := app.newTemplateData(r)
	userID := app.sessionManager.Get(r.Context(), "authenticatedUserID").(int)

	groups, err := app.models.Groups.GetAllForUser(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	templateData.Groups = groups

	for _, group := range groups {
		if group.Members == nil {
			continue

		}
		for _, member := range group.Members {
			app.infoLog.Println(member)
		}
	}

	app.render(w, http.StatusOK, "groups.tmpl", templateData)
}

func (app *application) viewOneGroupPage(w http.ResponseWriter, r *http.Request) {
	templateData := app.newTemplateData(r)
	userID := app.sessionManager.Get(r.Context(), "authenticatedUserID").(int)

	groupID, err := app.readIDParam(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest, "Invalid group ID in URL")
		return
	}

	group, err := app.models.Groups.Get(int(groupID))
	if err != nil {
		// app.serverError(w, err)
		app.clientError(w, http.StatusNotFound, "Group not found")
		return
	}

	// Check if user is in group.Members
	isMember := false
	for _, member := range group.Members {
		if member.ID == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		app.clientError(w, http.StatusForbidden, "You are not a member of this group")
		return
	}

	if group == nil {
		app.clientError(w, http.StatusNotFound, "Group not found")
		return
	}

	templateData.Group = group
	app.render(w, http.StatusOK, "group.tmpl", templateData)
}

// include struct tags to tell the decoder how to map HTML form vals to
// struct fields. "-" tells it to ignore a field!
type createGroupForm struct {
	Name                string `form:"name"`
	Description         string `form:"description"`
	validator.Validator `form:"-"`
}

func (app *application) createGroup(w http.ResponseWriter, r *http.Request) {
	userID := app.sessionManager.Get(r.Context(), "authenticatedUserID").(int)
	var form createGroupForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	// Validate the form.
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Name, 50), "name", "This field is too long")
	form.CheckField(validator.NotBlank(form.Description), "description", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Description, 1000), "description", "This field is too long")

	if !form.Valid() {
		app.clientError(w, http.StatusUnprocessableEntity, "Invalid form data")
		return
	}

	_, err = app.models.Groups.Insert(userID, form.Name, form.Description)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Group created successfully!")
	http.Redirect(w, r, "/groups", http.StatusSeeOther)
}

type inviteUserForm struct {
	Email               string `form:"email"`
	validator.Validator `form:"-"`
}

func (app *application) inviteUserToGroup(w http.ResponseWriter, r *http.Request) {
	currUserID := app.sessionManager.Get(r.Context(), "authenticatedUserID").(int)
	groupID, err := app.readIDParam(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest, "Invalid group ID in URL")
		return
	}

	var form inviteUserForm
	err = app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	// Get the group by ID.
	group, err := app.models.Groups.Get(int(groupID))
	// Check if the currUser is a member of the group.
	isMember := false
	for _, member := range group.Members {
		if member.ID == currUserID {
			isMember = true
			break
		}
	}
	if !isMember {
		app.clientError(w, http.StatusForbidden, "You are not a member of this group, can't invite!")
		return
	}

	// Get the user by email.
	user, err := app.models.Users.GetByEmail(form.Email)
	err = app.models.Groups.AddMember(int(groupID), user.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "User invited to group successfully!")
	http.Redirect(w, r, "/groups", http.StatusSeeOther)

}
