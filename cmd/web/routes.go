package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	mux := pat.New()
	mux.Get("/", http.HandlerFunc(app.home))
	mux.Get("/maintenanceRequest/create", http.HandlerFunc(app.createMaintenanceRequestForm))
	mux.Post("/maintenanceRequest/create", http.HandlerFunc(app.createMaintenanceRequest))
	mux.Get("/maintenanceRequest/:id", http.HandlerFunc(app.showMaintenanceRequest))

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	return standardMiddleware.Then(mux)
}
