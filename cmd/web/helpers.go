package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/go-playground/form/v4"
)

// serverError writes the stack error message.
// It sends a generic 500 internal server error to the user
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	if app.debug {
		http.Error(w, trace, http.StatusInternalServerError)
		return
	}

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// clientError sends a given status code to the user
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// notFound is a wrapper around clientError for the status 404 Not Found
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

// render will use the in memory cached template and execute the given page
func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exit", page)
		app.serverError(w, err)
		return
	}

	// trial render to check for runtime errors
	b := &bytes.Buffer{}
	err := ts.ExecuteTemplate(b, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.WriteHeader(status)

	b.WriteTo(w)
}

// decodePostForm will parse the requested form to a given destination
func (app *application) decodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		return err
	}

	return nil
}

// isAuthenticated returns true if a user is logged in
func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}

	return isAuthenticated
}
