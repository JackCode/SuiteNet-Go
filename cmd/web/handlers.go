package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackcode/suitenet/pkg/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Retrieve Incomplete Maintenance Requests to display on home page
	mr, err := app.maintenanceRequests.OpenPendingInProgress()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "home.page.tmpl", &templateData{
		MaintenanceRequests: mr,
	})
}

func (app *application) showMaintenanceRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	mr, err := app.maintenanceRequests.Get(id)
	if err == models.ErrNoRecord {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "show.page.tmpl", &templateData{
		MaintenanceRequest: mr,
	})
}

func (app *application) maintenanceRequestForm(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Create a new snippet..."))
}

func (app *application) createMaintenanceRequest(w http.ResponseWriter, r *http.Request) {
	title := "Pool - Seresco not working"
	description := "Compressor 1 High Pressure fault"
	status := "OPEN"

	id, err := app.maintenanceRequests.Insert(title, description, status)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/maintenanceRequest/%d", id), http.StatusSeeOther)
}
