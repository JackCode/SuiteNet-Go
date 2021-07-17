package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	dynamicMiddleware := alice.New(app.session.Enable, noSurf, app.authenticate)

	mux := pat.New()
	// Dashboard
	mux.Get("/", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.dashboard))

	// Request Routes
	mux.Get("/:department/request/create", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.createRequestForm))
	mux.Post("/:department/request/create", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.createRequest))
	mux.Get("/:department/request/all", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.allRequests))
	mux.Get("/:department/request/incomplete", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.showIncompleteRequests))
	mux.Get("/:department/request/:id", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.showRequest))
	mux.Post("/:department/request/:id/close", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.closeRequest))
	mux.Post("/:department/request/:id/reopen", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.reopenRequest))
	mux.Post("/:department/request/:id/addNote", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.addNoteToRequest))

	// User routes
	mux.Get("/user/signup", dynamicMiddleware.Append(app.requireRoles("admin")).ThenFunc(app.signupUserForm))
	mux.Post("/user/signup", dynamicMiddleware.Append(app.requireRoles("admin")).ThenFunc(app.signupUser))
	mux.Get("/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm))
	mux.Post("/user/login", dynamicMiddleware.ThenFunc(app.loginUser))
	mux.Post("/user/logout", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.logoutUser))
	mux.Get("/user/resetPassword", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.resetPasswordForm))
	mux.Post("/user/resetPassword", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.resetPassword))
	mux.Get("/user/clock/:direction", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.clockUser))

	// Miscellaneous Routes

	// Static server
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	return standardMiddleware.Then(mux)
}
