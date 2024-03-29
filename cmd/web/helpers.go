package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/jackcode/suitenet/pkg/models"
	"github.com/justinas/nosurf"
)

// The serverError helper writes an error message and stack trace to the errorLog,
// then send a generic 500 Internal Server Error response to the header
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// The clientError helper sends a specific status code and corresponding description
// to the user. We'll use this later in the book to send responses like 400 "Bad
// Request" when there's a problem with the request that the user sent.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// For consistency, we'll also implement a notFound helper. This is simply a
// convenience wrapper around clientError which sends a 404 Not Found response to
// the user.
func (app *application) notFound(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "not_found.page.tmpl", &templateData{})
}

func (app *application) accessDenied(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "denied.page.tmpl", &templateData{})
}

// Create an addDefaultData helper. This takes a pointer to a templateData
// struct, adds the current year to the CurrentYear field, and then returns
// the pointer. Again, we're not using the *http.Request parameter at the
// moment, but we will do later in the book.
func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	if td == nil {
		td = &templateData{}
	}

	locations, err := app.locations.GetActiveLocations()
	if err == nil {
		td.Locations = locations
	}

	positions, err := app.positions.GetActivePositions()
	if err == nil {
		td.Positions = positions
	}

	users, err := app.sys_users.GetActiveUsers()
	if err == nil {
		td.Users = users
	}

	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	td.Flash = app.session.PopString(r, "flash")
	td.AuthenticatedUser = app.authenticatedUser(r)
	return td
}

func (app *application) render(w http.ResponseWriter, r *http.Request, name string, td *templateData) {
	// Retrieve the appropriate template set from the cache based on the page name
	// (like 'home.page.tmpl'). If no entry exists in the cache with the
	// provided name, call the serverError helper method that we made earlier.
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	buf := new(bytes.Buffer)

	// Execute the template set, passing in any dynamic data.
	err := ts.Execute(buf, app.addDefaultData(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}

	buf.WriteTo(w)
}

// The authenticatedUser method returns the ID of the current user from the
// session, or zero if the request is from an unauthenticated user.
func (app *application) authenticatedUser(r *http.Request) *models.SysUser {
	user, ok := r.Context().Value(contextKeyUser).(*models.SysUser)
	if !ok {
		return nil
	}
	return user
}

// The authenticatedUser method returns the ID of the current user from the
// session, or zero if the request is from an unauthenticated user.
func (app *application) authenticatedRole(r *http.Request, requiredRole string) *models.SysUser {
	user, ok := r.Context().Value(contextKeyUser).(*models.SysUser)
	if !ok || !rolesContain(user.SiteRoles, requiredRole) {
		return nil
	}
	return user
}
