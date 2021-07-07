package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackcode/suitenet/pkg/forms"
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

func (app *application) createMaintenanceRequestForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "create.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func (app *application) createMaintenanceRequest(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("title", "description", "status")
	form.MaxLength("title", 100)
	form.PermittedValues("status", "OPEN", "PENDING", "IN PROGRESS", "COMPLETE")

	if !form.Valid() {
		app.render(w, r, "create.page.tmpl", &templateData{Form: form})
		return
	}

	id, err := app.maintenanceRequests.Insert(form.Get("title"), form.Get("description"), form.Get("status"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/maintenanceRequest/%d", id), http.StatusSeeOther)
}
