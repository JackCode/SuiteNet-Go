package main

import (
	"fmt"
	"net/http"	"runtime/debug"
)

// The serverError helper writes an error message and stack trace to the errorLog,
// then send a generic 500 Internal Server Error response to the header
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Println(trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

