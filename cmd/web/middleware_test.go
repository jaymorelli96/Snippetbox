package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"snippetbox.jmorelli.dev/internal/assert"
)

func TestSecureHeaders(t *testing.T) {
	rr := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	secureHeaders(next).ServeHTTP(rr, r)

	res := rr.Result()
	defer res.Body.Close()

	expectedValue := "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com"
	assert.Equal(t, res.Header.Get("Content-Security-Policy"), expectedValue)

	expectedValue = "origin-when-cross-origin"
	assert.Equal(t, res.Header.Get("Referrer-Policy"), expectedValue)

	expectedValue = "nosniff"
	assert.Equal(t, res.Header.Get("X-Content-Type-Options"), expectedValue)

	expectedValue = "deny"
	assert.Equal(t, res.Header.Get("X-Frame-Options"), expectedValue)

	expectedValue = "0"
	assert.Equal(t, res.Header.Get("X-XSS-Protection"), expectedValue)

	assert.Equal(t, res.StatusCode, http.StatusOK)

	assert.NextHandler(t, res.Body)
}

func TestLogRequest(t *testing.T) {
	rr := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "0.0.0.0"

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	buf := bytes.Buffer{}

	app := &application{
		infoLog:  log.New(&buf, "INFO ", 0),
		errorLog: log.New(io.Discard, "", 0),
	}

	app.logRequest(next).ServeHTTP(rr, r)

	if buf.String() != "INFO 0.0.0.0 - HTTP/1.1 GET /\n" {
		t.Errorf("got %q; want %q", buf.String(), "INFO 0.0.0.0 - HTTP/1.1 GET /\n")
	}

	res := rr.Result()
	defer res.Body.Close()

	assert.NextHandler(t, res.Body)
}

func TestRequireAuthentication(t *testing.T) {

	tests := []struct {
		name          string
		authenticated bool
		headerCache   string
		nextHandler   bool
	}{
		{
			name:          "Authenticated User",
			authenticated: true,
			headerCache:   "no-store",
			nextHandler:   true,
		},
		{
			name:          "Non-Authenticated User",
			authenticated: false,
			headerCache:   "",
			nextHandler:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, tt.authenticated)
			r = r.WithContext(ctx)

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("OK"))
			})

			app := &application{
				infoLog:  log.New(io.Discard, "", 0),
				errorLog: log.New(io.Discard, "", 0),
			}

			app.requireAuthentication(next).ServeHTTP(rr, r)

			res := rr.Result()
			defer res.Body.Close()

			assert.Equal(t, res.Header.Get("Cache-Control"), tt.headerCache)

			if tt.nextHandler {
				assert.NextHandler(t, res.Body)
			}

		})
	}
}
