package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogReq_DelegatesAndLogs(t *testing.T) {
	var called bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	buf := &bytes.Buffer{}
	prevFlags := log.Flags()
	prevOutput := log.Writer()
	log.SetFlags(0)
	log.SetOutput(buf)
	defer log.SetFlags(prevFlags)
	defer log.SetOutput(prevOutput)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/demo", nil)

	LogReq(handler).ServeHTTP(rr, req)

	if !called {
		t.Fatalf("expected wrapped handler to be called")
	}
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
	if got := buf.String(); got == "" || !bytes.Contains(buf.Bytes(), []byte("GET /demo")) {
		t.Fatalf("expected log output to contain method/path, got %q", got)
	}
}
