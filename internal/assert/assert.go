package assert

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func Equal[T comparable](t testing.TB, actual, expected T) {
	t.Helper()

	if actual != expected {
		t.Errorf("got: %v; want: %v", actual, expected)
	}
}

func StringContains(t testing.TB, actual, expectedSubs string) {
	t.Helper()

	if !strings.Contains(actual, expectedSubs) {
		t.Errorf("got: %q; expected to contain: %q", actual, expectedSubs)
	}
}

func NilError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Errorf("got: %v; expected: nil", actual)
	}
}

func NextHandler(t testing.TB, body io.Reader) {
	b, err := io.ReadAll(body)
	if err != nil {
		t.Fatal(err)
	}

	bytes.TrimSpace(b)

	Equal(t, string(b), "OK")
}
