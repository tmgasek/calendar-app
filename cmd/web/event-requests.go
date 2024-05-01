package main

import (
	"fmt"
	"net/http"
)

func (app *application) updateRequestStatus(w http.ResponseWriter, r *http.Request) {

	eventID, err := app.readIDParam(r)
	if err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	fmt.Printf("eventID: %v\n", eventID)

	action := r.FormValue("action")
	//TODO: validate action to be one of "confirmed" or "declined"

	if action != "confirmed" && action != "declined" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if action == "declined" {
		// TODO: delete the event from db
		app.infoLog.Printf("Event %d declined\n", eventID)
		return
	}

	// Accepted at this point, need to:
	// - Get the event from DB
	// - Using its info, add the events to each of user's linked calendars
	// - Remove the pending event from DB
	// - Refetch the events from each calendar into events table
	http.Redirect(w, r, "/requests", http.StatusSeeOther)
}
