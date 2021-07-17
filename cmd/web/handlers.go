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

func (app *application) showIncompleteRequests(w http.ResponseWriter, r *http.Request) {
	department := r.URL.Query().Get(":department")
	request, err := app.requests.GetIncompleteRequests(department)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "incomplete_requests.page.tmpl", &templateData{
		Requests:   request,
		Department: department,
	})
}

func (app *application) allRequests(w http.ResponseWriter, r *http.Request) {
	department := r.URL.Query().Get(":department")
	requests, err := app.requests.GetAllRequests(department)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "all_requests.page.tmpl", &templateData{
		Requests:   requests,
		Department: department,
	})
}

func (app *application) showRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w, r)
		return
	}

	department := r.URL.Query().Get(":department")
	request, err := app.requests.Get(id, department)
	if err == models.ErrNoRecord {
		app.notFound(w, r)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}
	err = app.requests.MarkRead(id, app.session.GetInt(r, "userID"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "show.page.tmpl", &templateData{
		Request:    request,
		Department: department,
	})
}

func (app *application) createRequestForm(w http.ResponseWriter, r *http.Request) {
	department := r.URL.Query().Get(":department")
	app.render(w, r, "create.page.tmpl", &templateData{
		Form:       forms.New(nil),
		Department: department,
	})
}

func (app *application) createRequest(w http.ResponseWriter, r *http.Request) {
	department := r.URL.Query().Get(":department")
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("title", "location")
	form.MaxLength("title", 100)

	if !form.Valid() {
		app.render(w, r, "create.page.tmpl", &templateData{
			Form:       form,
			Department: department})
		return
	}

	id, err := app.requests.Insert(form.Get("title"), form.Get("location"), form.Get("note"), department, app.session.GetInt(r, "userID"))
	if id == 0 {
		app.session.Put(r, "flash", "Internal error creating request.")
		app.render(w, r, "create.page.tmpl", &templateData{
			Form:       form,
			Department: department,
		})
		return
	}
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r, "flash", "Request successfully created!")
	http.Redirect(w, r, fmt.Sprintf("/%s/request/%d", department, id), http.StatusSeeOther)
}

func (app *application) closeRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w, r)
		return
	}
	department := r.URL.Query().Get(":department")
	redirectURL := fmt.Sprintf("/%s/request/%d", department, id)

	_, err = app.requests.Close(id, app.session.GetInt(r, "userID"), department)
	if err == models.ErrNoRecord {
		http.Redirect(w, r, redirectURL, 303)
		app.session.Put(r, "generic", "Request not found.")
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r, "flash", "Request closed successfully.")

	http.Redirect(w, r, redirectURL, 303)
}

func (app *application) reopenRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w, r)
		return
	}
	department := r.URL.Query().Get(":department")
	redirectURL := fmt.Sprintf("/%s/request/%d", department, id)

	_, err = app.requests.Reopen(id, app.session.GetInt(r, "userID"), department)
	if err == models.ErrNoRecord {
		http.Redirect(w, r, redirectURL, 303)
		app.session.Put(r, "generic", "Request not found.")
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, redirectURL, 303)
	app.session.Put(r, "flash", "Request reopened successfully.")

	http.Redirect(w, r, redirectURL, 303)
}

func (app *application) addNoteToRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w, r)
		return
	}
	department := r.URL.Query().Get(":department")
	redirectURL := fmt.Sprintf("/%s/request/%d", department, id)

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

	_, err = app.requests.AddNote(form.Get("note"), department, id, app.session.GetInt(r, "userID"))
	if err == models.ErrNoRecord {
		app.session.Put(r, "flash", "Request not found.")
		http.Redirect(w, r, redirectURL, 303)
		return
	} else if err != nil {
		app.serverError(w, err)
		http.Redirect(w, r, redirectURL, 303)
		return
	}

	app.session.Put(r, "flash", "Note added to request successfully.")

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
