package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackcode/suitenet/pkg/forms"
	"github.com/jackcode/suitenet/pkg/models"
)

func (app *application) dashboard(w http.ResponseWriter, r *http.Request) {

	app.render(w, r, "dashboard.page.tmpl", &templateData{})
}

func (app *application) engineering(w http.ResponseWriter, r *http.Request) {
	// Retrieve Incomplete Maintenance Requests to display on home page
	mr, err := app.workOrders.OpenPendingInProgress()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "engineering.page.tmpl", &templateData{
		WorkOrders: mr,
	})
}

func (app *application) showWorkOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	mr, err := app.workOrders.Get(id)
	if err == models.ErrNoRecord {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "show.page.tmpl", &templateData{
		WorkOrder: mr,
	})
}

func (app *application) createWorkOrderForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "create.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func (app *application) createWorkOrder(w http.ResponseWriter, r *http.Request) {
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

	id, err := app.workOrders.Insert(form.Get("title"), form.Get("description"), form.Get("status"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r, "flash", "Maintenance Request successfully created!")
	http.Redirect(w, r, fmt.Sprintf("/engineering/workOrder/%d", id), http.StatusSeeOther)
}

func (app *application) signupUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "signup.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func (app *application) signupUser(w http.ResponseWriter, r *http.Request) {
	// Parse the form data.
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the form contents using the form helper we made earlier.
	form := forms.New(r.PostForm)
	form.Required("name", "username", "password")
	form.MinLength("password", 8)
	form.MinLength("name", 2)
	form.MinLength("username", 4)

	// If there are any errors, redisplay the signup form.
	if !form.Valid() {
		app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		return
	}

	// Try to create a new user record in the database. If the username already exists
	// add an error message to the form and re-display it.
	err = app.users.Insert(form.Get("name"), form.Get("username"), form.Get("password"))
	if err == models.ErrDuplicateUsername {
		form.Errors.Add("username", "Username is already in use")
		app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	// Otherwise add a confirmation flash message to the session confirming that
	// their signup worked and asking them to log in.
	app.session.Put(r, "flash", "Signup was successful.")

	// And redirect the user to the login page.
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) loginUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Check whether the credentials are valid. If they're not, add a generic error
	// message to the form failures map and re-display the login page.
	form := forms.New(r.PostForm)
	id, err := app.users.Authenticate(form.Get("username"), form.Get("password"))
	if err == models.ErrInvalidCredentials {
		form.Errors.Add("generic", "Username or Password is incorrect")
		app.render(w, r, "login.page.tmpl", &templateData{Form: form})
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	// Add the ID of the current user to the session, so that they are now 'logged
	// in'.
	app.session.Put(r, "userID", id)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {
	// Remove the userID from the session data so that the user is 'logged out'.
	app.session.Remove(r, "userID")
	// Add a flash message to the session to confirm to the user that they've been logged out.
	app.session.Put(r, "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", 303)
}
