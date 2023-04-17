package main

import (
	"bytes"
	"html"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"snippetbox.jmorelli.dev/internal/models/mocks"
)

func newTestApplication(t *testing.T) *application {
	templateCache, err := newTemplateCache()
	if err != nil {
		t.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	return &application{
		errorLog:       log.New(io.Discard, "", 0),
		infoLog:        log.New(io.Discard, "", 0),
		snippets:       &mocks.SnippetModel{},
		users:          &mocks.UserModel{},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}
}

type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	srv := httptest.NewTLSServer(h)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	srv.Client().Jar = jar

	srv.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{srv}
}

func (srv *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	rs, err := srv.Client().Get(srv.URL + urlPath)

	if err != nil {
		t.Fatal(err)
	}

	body, err := io.ReadAll(rs.Body)
	defer rs.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	body = bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}

func (srv *testServer) post(t *testing.T, urlPath string, form url.Values) (int, http.Header, string) {
	rs, err := srv.Client().PostForm(srv.URL+urlPath, form)
	if err != nil {
		t.Fatal(err)
	}

	body, err := io.ReadAll(rs.Body)
	defer rs.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	body = bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}

var csrfTokenRX = regexp.MustCompile(`<input type='hidden' name='csrf_token' value='(.+)'>`)

func extractCSRFToken(t testing.TB, body string) string {
	matches := csrfTokenRX.FindStringSubmatch(body)
	if matches == nil {
		t.Fatal("no CRSF token found in body")
	}

	return html.UnescapeString(string(matches[1]))
}
