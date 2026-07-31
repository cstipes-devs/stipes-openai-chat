package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_OK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	rr := httptest.NewRecorder()

	Handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want application/json", got)
	}
	if got, want := rr.Body.String(), `{"ok":true}`; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}
