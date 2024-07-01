package main

import "net/http"

func (app *application) getLocalStateHandler(w http.ResponseWriter, r *http.Request) {
	groups, err := app.hue.Bridge.GetGroups()
	app.groups = &groups
	if err != nil {
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusInternalServerError, "Failed to get groups")
		return
	}
	app.writeJSON(w, http.StatusOK, envelope{"state": groups}, nil)
}

// func (a *application) getRemoteState(w http.ResponseWriter, r *http.Request) {
// 	groups, err := a.hue.Bridge.GetGroups()
// 	if err != nil {
// 		a.logger.Error(err.Error())
// 		a.errorResponse(w, r, http.StatusInternalServerError, "Failed to get state")
// 		return
// 	}
// 	groups
// 	a.writeJSON(w, http.StatusOK, envelope{"state": state}, nil)
// }
