package openldap_exporter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	// tag and commit are injected via ldflags; in tests they are empty strings.
	v := GetVersion()
	if !strings.Contains(v, "(") || !strings.Contains(v, ")") {
		t.Errorf("unexpected version format: %q", v)
	}
}

func TestShowVersion_GET(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()
	showVersion(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("Content-Type: got %q, want text/plain", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, GetVersion()) {
		t.Errorf("body %q does not contain version %q", body, GetVersion())
	}
}

func TestShowVersion_MethodNotAllowed(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		req := httptest.NewRequest(method, "/version", nil)
		rec := httptest.NewRecorder()
		showVersion(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: got %d, want %d", method, rec.Code, http.StatusMethodNotAllowed)
		}
	}
}
