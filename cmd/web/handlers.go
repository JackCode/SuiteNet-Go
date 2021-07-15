package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackcode/suitenet/pkg/forms"
	"github.com/jackcode/suitenet/pkg/models"
)

func (app *application) dashboard(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "dashboard.page.tmpl", &templateData{})
}

func (app *application) engineering(w http.ResponseWriter, r *http.Request) {
	// Retrieve Incomplete Maintenance Requests to display on home page
	mr, err := app.requests.GetIncompleteRequests()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "engineering.page.tmpl", &templateData{
		Requests: mr,
	})
}

func (app *application) allWorkOrders(w http.ResponseWriter, r *http.Request) {
	// Retrieve Incomplete Maintenance Requests to display on home page
	mr, err := app.requests.GetAllWorkOrders()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "engineering.all.page.tmpl", &templateData{
		Requests: mr,
	})
}

func (app *application) showWorkOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	workOrder, err := app.requests.Get(id)
	if err == models.ErrNoRecord {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}
	err = app.requests.Read(id, app.session.GetInt(r, "userID"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "show.page.tmpl", &templateData{
		Request: workOrder,
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
	form.Required("title", "location")
	form.MaxLength("title", 100)

	if !form.Valid() {
		app.render(w, r, "create.page.tmpl", &templateData{Form: form})
		return
	}

	app.infoLog.Printf("Creating work oder: Title: %s, Location: %s, Content: %s, UserID: %d", form.Get("title"), form.Get("location"), form.Get("note"), app.session.GetInt(r, "userID"))
	id, err := app.requests.Insert(form.Get("title"), form.Get("location"), form.Get("note"), app.session.GetInt(r, "userID"))
	if id == 0 {
		app.session.Put(r, "flash", "Internal error creating work order.")
		app.render(w, r, "create.page.tmpl", &templateData{Form: form})
		return
	}
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r, "flash", "Maintenance work order successfully created!")
	http.Redirect(w, r, fmt.Sprintf("/engineering/workOrder/%d", id), http.StatusSeeOther)
}

func (app *application) closeWorkOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	redirectURL := fmt.Sprintf("/engineering/workOrder/%d", id)

	_, err = app.requests.Close(id, app.session.GetInt(r, "userID"))
	if err == models.ErrNoRecord {
		http.Redirect(w, r, redirectURL, 303)
		app.session.Put(r, "generic", "Work order not found.")
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r, "flash", "Work order closed successfully.")

	http.Redirect(w, r, redirectURL, 303)
}

func (app *application) reopenWorkOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	redirectURL := fmt.Sprintf("/engineering/workOrder/%d", id)

	_, err = app.requests.Reopen(id, app.session.GetInt(r, "userID"))
	if err == models.ErrNoRecord {
		http.Redirect(w, r, redirectURL, 303)
		app.session.Put(r, "generic", "Work order not found.")
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, redirectURL, 303)
	app.session.Put(r, "flash", "Work order reopened successfully.")

	http.Redirect(w, r, redirectURL, 303)
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
	form.Required("name", "username", "password", "position", "manager")
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
	err = app.sys_users.Insert(strings.TrimSpace(form.Get("name")),
		strings.TrimSpace(form.Get("username")),
		form.Get("password"),
		form.Get("position"),
		form.Get("manager"),
		app.session.GetInt(r, "userID"))

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
	id, err := app.sys_users.Authenticate(form.Get("username"), form.Get("password"))
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

func (app *application) resetPasswordForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "resetpassword.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func (app *application) resetPassword(w http.ResponseWriter, r *http.Request) {
	// Parse the form data.
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the form contents using the form helper we made earlier.
	form := forms.New(r.PostForm)
	form.Required("username", "password")
	form.MinLength("password", 8)
	form.MinLength("username", 4)

	// If there are any errors, redisplay the reset password form.
	if !form.Valid() {
		app.render(w, r, "resetpassword.page.tmpl", &templateData{Form: form})
		return
	}

	// Try to create a new user record in the database. If the username already exists
	// add an error message to the form and re-display it.
	err = app.sys_users.UpdatePassword(form.Get("username"), form.Get("password"))
	if err == models.ErrInvalidCredentials {
		form.Errors.Add("generic", "Error updating password. Please check username.")
		app.render(w, r, "resetpassword.page.tmpl", &templateData{Form: form})
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	// Otherwise add a confirmation flash message to the session confirming that
	// their signup worked and asking them to log in.
	app.session.Put(r, "flash", "Password update successful.")

	// And redirect the user to the login page.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) addNoteToWorkOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	redirectURL := fmt.Sprintf("/engineering/workOrder/%d", id)

	// Parse the form data.
	err = r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("note")
	// If there are any errors, redisplay the signup form.
	if !form.Valid() {
		http.Redirect(w, r, redirectURL, 303)
		app.session.Put(r, "flash", "Note field blank. No note added.")
		return
	}

	_, err = app.requests.AddNote(form.Get("note"), id, app.session.GetInt(r, "userID"))
	if err == models.ErrNoRecord {
		app.session.Put(r, "flash", "Work order not found.")
		http.Redirect(w, r, redirectURL, 303)
		return
	} else if err != nil {
		app.serverError(w, err)
		http.Redirect(w, r, redirectURL, 303)
		return
	}

	app.session.Put(r, "flash", "Note added to work order successfully.")

	http.Redirect(w, r, redirectURL, 303)
}
