package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"testing"

	"snippetbox.jmorelli.dev/internal/assert"
	"snippetbox.jmorelli.dev/internal/models"
)

func TestSnippetView(t *testing.T) {
	app := newTestApplication(t)

	srv := newTestServer(t, app.routes())
	defer srv.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{
			name:     "Valid ID",
			urlPath:  "/snippet/view/1",
			wantCode: http.StatusOK,
			wantBody: "An old silent pond...",
		},
		{
			name:     "Non-existent ID",
			urlPath:  "/snippet/view/2",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Decimal ID",
			urlPath:  "/snippet/view/3.14",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "String ID",
			urlPath:  "/snippet/view/test",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Empty ID",
			urlPath:  "/snippet/view/",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := srv.get(t, tt.urlPath)

			assert.Equal(t, code, tt.wantCode)

			if tt.wantBody != "" {
				assert.StringContains(t, body, tt.wantBody)
			}
		})
	}
}

func TestUserSignup(t *testing.T) {
	app := newTestApplication(t)

	srv := newTestServer(t, app.routes())
	defer srv.Close()

	_, _, body := srv.get(t, "/user/signup")
	validCSRFToken := extractCSRFToken(t, body)

	const (
		validName     = "John"
		validPassword = "12345678"
		validEmail    = "johndoe@email.com"
		formTag       = "<form action='/user/signup' method='POST' novalidate>"
	)

	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantFormTag  string
	}{
		{
			name:         "Valid submission",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusSeeOther,
		},
		{
			name:         "Invalid CSRF Token",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    "wrongToken",
			wantCode:     http.StatusBadRequest,
		},
		{
			name:         "Empty name",
			userName:     "",
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Empty email",
			userName:     validName,
			userEmail:    "",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Empty password",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: "",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Invalid email",
			userName:     validName,
			userEmail:    "bob@example.",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Short password",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: "pa$$",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Duplicate email",
			userName:     validName,
			userEmail:    "dupe@example.com",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("name", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)

			code, _, body := srv.post(t, "/user/signup", form)

			assert.Equal(t, code, tt.wantCode)

			if tt.wantFormTag != "" {
				assert.StringContains(t, body, tt.wantFormTag)
			}
		})
	}
}

func TestHome(t *testing.T) {
	app := newTestApplication(t)
	appError := &application{
		infoLog:  log.New(io.Discard, "", 0),
		errorLog: log.New(io.Discard, "", 0),
		snippets: &models.SnippetModel{},
	}

	tests := []struct {
		name     string
		urlPath  string
		app      *application
		wantCode int
		wantBody string
	}{
		{
			name:     "Home page",
			urlPath:  "/",
			app:      app,
			wantCode: http.StatusOK,
			wantBody: "Latest Snippets",
		},
		{
			name:     "Show latest snippets",
			urlPath:  "/",
			app:      app,
			wantCode: http.StatusOK,
			wantBody: "An old silent pond",
		},
		{
			name:     "Internal Error",
			urlPath:  "/",
			app:      appError,
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newTestServer(t, tt.app.routes())
			defer srv.Close()

			code, _, body := srv.get(t, tt.urlPath)

			assert.Equal(t, code, tt.wantCode)

			if tt.wantBody != "" {
				assert.StringContains(t, body, tt.wantBody)
			}
		})
	}
}

func TestSnippetCreate(t *testing.T) {
	app := newTestApplication(t)

	srv := newTestServer(t, app.routes())
	defer srv.Close()

	t.Run("Unauthenticated", func(t *testing.T) {
		code, headers, _ := srv.get(t, "/snippet/create")

		assert.Equal(t, code, http.StatusSeeOther)
		assert.Equal(t, headers.Get("Location"), "/user/login")
	})

	t.Run("Authenticated", func(t *testing.T) {
		// Make a GET /user/login request and extract the CSRF token from the
		// response.
		_, _, body := srv.get(t, "/user/login")
		csrfToken := extractCSRFToken(t, body)

		// Make a POST /user/login request using the extracted CSRF token and
		// credentials from our the mock user model.
		form := url.Values{}
		form.Add("email", "jay@email.com")
		form.Add("password", "12345678")
		form.Add("csrf_token", csrfToken)
		srv.post(t, "/user/login", form)

		// Then check that the authenticated user is shown the create snippet
		// form.
		code, _, body := srv.get(t, "/snippet/create")

		assert.Equal(t, code, http.StatusOK)
		assert.StringContains(t, body, "<form action='/snippet/create' method='POST'>")
	})
}
