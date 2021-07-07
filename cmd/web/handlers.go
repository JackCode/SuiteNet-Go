package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackcode/suitenet/pkg/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	mr, err := app.maintenanceRequests.OpenAndPending()
	if err != nil {
		app.serverError(w, err)
		return
	}

	for _, request := range mr {
		fmt.Fprintf(w, "%v\n\n", request)
	}

	// // Include the footer partial in the template files.
	// files := []string{
	// 	"./ui/html/home.page.tmpl",
	// 	"./ui/html/base.layout.tmpl",
	// 	"./ui/html/footer.partial.tmpl",
	// }

	// ts, err := template.ParseFiles(files...)
	// if err != nil {
	// 	app.serverError(w, err)
	// 	return
	// }

	// err = ts.Execute(w, nil)
	// if err != nil {
	// 	app.serverError(w, err)
	// }
}

func (app *application) showMaintenanceRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
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

	fmt.Fprintf(w, "%v", mr)
}

func (app *application) createMaintenanceRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	title := "Pool - Seresco not working"
	description := "Compressor 1 High Pressure fault"
	status := "OPEN"

	id, err := app.maintenanceRequests.Insert(title, description, status)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/maintenanceRequest?id=%d", id), http.StatusSeeOther)
}
