package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"snippetbox.jmorelli.dev/ui"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(http.FS(ui.Files))
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	router.Handler(
		http.MethodGet, "/",
		app.sessionManager.LoadAndSave(
			app.authenticate(http.HandlerFunc(app.home)),
		),
	)
	router.Handler(
		http.MethodGet, "/about",
		app.sessionManager.LoadAndSave(
			app.authenticate(http.HandlerFunc(app.about)),
		),
	)
	router.Handler(
		http.MethodGet, "/snippet/view/:id",
		app.sessionManager.LoadAndSave(
			app.authenticate(http.HandlerFunc(app.snippetView)),
		),
	)
	router.Handler(
		http.MethodGet, "/snippet/create",
		app.sessionManager.LoadAndSave(
			app.authenticate(app.requireAuthentication(http.HandlerFunc(app.snippetCreate))),
		),
	)
	router.Handler(
		http.MethodPost, "/snippet/create",
		app.sessionManager.LoadAndSave(
			app.authenticate(app.requireAuthentication(http.HandlerFunc(app.snippetCreatePost))),
		),
	)
	router.Handler(
		http.MethodGet, "/user/signup",
		app.sessionManager.LoadAndSave(
			app.authenticate(http.HandlerFunc(app.userSignup)),
		),
	)
	router.Handler(
		http.MethodPost, "/user/signup", app.sessionManager.LoadAndSave(
			app.authenticate(http.HandlerFunc(app.userSignupPost)),
		),
	)
	router.Handler(
		http.MethodGet, "/user/login",
		app.sessionManager.LoadAndSave(
			app.authenticate(http.HandlerFunc(app.userLogin)),
		),
	)
	router.Handler(
		http.MethodPost, "/user/login",
		app.sessionManager.LoadAndSave(
			app.authenticate(http.HandlerFunc(app.userLoginPost)),
		),
	)
	router.Handler(
		http.MethodPost, "/user/logout",
		app.sessionManager.LoadAndSave(
			app.authenticate(app.requireAuthentication(http.HandlerFunc(app.userLogoutPost))),
		),
	)
	router.Handler(
		http.MethodGet, "/account/view",
		app.sessionManager.LoadAndSave(
			app.authenticate(app.requireAuthentication(http.HandlerFunc(app.userAccount))),
		),
	)
	router.Handler(
		http.MethodGet, "/account/password/update",
		app.sessionManager.LoadAndSave(
			app.authenticate(app.requireAuthentication(http.HandlerFunc(app.passwordUpdate))),
		),
	)
	router.Handler(
		http.MethodPost, "/account/password/update",
		app.sessionManager.LoadAndSave(
			app.authenticate(app.requireAuthentication(http.HandlerFunc(app.passwordUpdatePost))),
		),
	)
	return app.recoverFromPanic(app.logRequest(app.noSurf(secureHeaders(router))))
}
