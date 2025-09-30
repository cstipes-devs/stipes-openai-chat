package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/chris-stipes/stipes-openai-chat/pkg/easter"
)

func TestFilesHandler_ReturnsInjectedMap(t *testing.T) {
	mux := http.NewServeMux()

	files := map[string]string{
		"stipes.md": "hello world",
		"resume.md": "# Chris Stipes\nSoftware engineer...",
	}

	Register(mux, Deps{
		Files:  files,
		APIKey: "test-key",
		Model:  "gpt-4o-mini",
	})

	req := httptest.NewRequest(http.MethodGet, "/files", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("content-type = %q, want application/json", ct)
	}

	var resp struct {
		Files map[string]string `json:"files"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !reflect.DeepEqual(resp.Files, files) {
		t.Fatalf("files mismatch\n got: %#v\nwant: %#v", resp.Files, files)
	}
}

func TestHealthz_OK(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, Deps{})

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("content-type = %q, want application/json", ct)
	}
	body := strings.TrimSpace(rr.Body.String())
	if body != "{\"ok\":true}" {
		t.Errorf("body = %q, want {\"ok\":true}", body)
	}
}

func TestChat_MethodNotAllowed(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, Deps{})

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/chat", nil))

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestChat_InvalidJSON(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, Deps{})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rr.Body.String(), "invalid json") {
		t.Errorf("expected invalid json error, got %q", rr.Body.String())
	}
}

func TestChat_OpenAIError(t *testing.T) {
	// Use empty API key to force the OpenAI layer to return an error
	// without making any network calls.
	mux := http.NewServeMux()
	Register(mux, Deps{APIKey: "", Model: "gpt-4o-mini"})

	payload := map[string]string{"message": "hello"}
	buf, _ := json.Marshal(payload)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadGateway)
	}
	body, _ := io.ReadAll(rr.Body)
	if !strings.Contains(string(body), "openai error:") {
		t.Errorf("expected openai error in body, got %q", string(body))
	}
}

func TestChat_EasterEgg(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, Deps{})

	payload := map[string]string{"message": "Please tell me an Easter secret"}
	buf, _ := json.Marshal(payload)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("content-type = %q, want application/json", ct)
	}

	var resp struct {
		Output string `json:"output"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	egg, _ := easter.Message("easter")
	if resp.Output != egg {
		t.Fatalf("output mismatch\n got: %q\nwant: %q", resp.Output, egg)
	}
}
