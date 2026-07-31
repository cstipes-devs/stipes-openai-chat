package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func withFiles(m map[string]string) func() {
	prev := files
	files = m
	return func() { files = prev }
}

func TestHandler_ReturnsFiles(t *testing.T) {
	defer withFiles(map[string]string{"resume.md": "# Resume"})()

	req := httptest.NewRequest(http.MethodGet, "/api/files", nil)
	rr := httptest.NewRecorder()

	Handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want application/json", got)
	}

	var resp struct {
		Files map[string]string `json:"files"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got, want := resp.Files["resume.md"], "# Resume"; got != want {
		t.Fatalf("files[resume.md] = %q, want %q", got, want)
	}
}
