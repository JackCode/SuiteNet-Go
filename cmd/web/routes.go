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

	// Engineering Routes
	mux.Get("/engineering/workOrder/create", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.createWorkOrderForm))
	mux.Post("/engineering/workOrder/create", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.createWorkOrder))
	mux.Get("/engineering/workOrder/:id", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.showWorkOrder))
	mux.Post("/engineering/workOrder/:id/close", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.closeWorkOrder))
	mux.Post("/engineering/workOrder/:id/reopen", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.reopenWorkOrder))
	mux.Get("/engineering", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.engineering))
	mux.Get("/engineering/workOrder", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.allWorkOrders))
	mux.Post("/engineering/workOrder/:id/addNote", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.addNoteToWorkOrder))

	// User routes
	mux.Get("/user/signup", dynamicMiddleware.ThenFunc(app.signupUserForm))
	mux.Post("/user/signup", dynamicMiddleware.ThenFunc(app.signupUser))
	mux.Get("/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm))
	mux.Post("/user/login", dynamicMiddleware.ThenFunc(app.loginUser))
	mux.Post("/user/logout", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.logoutUser))
	mux.Get("/user/resetPassword", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.resetPasswordForm))
	mux.Post("/user/resetPassword", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.resetPassword))

	// Static server
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	return standardMiddleware.Then(mux)
}
