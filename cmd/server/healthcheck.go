package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create a map which holds the information that we want to send in the response.
	data := envelope{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}
	if err := app.writeJSON(w, http.StatusOK, data, nil); err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}
