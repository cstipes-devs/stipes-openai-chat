package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	ai "github.com/chris-stipes/stipes-openai-chat/pkg/claude"
	"github.com/chris-stipes/stipes-openai-chat/pkg/easter"
)

func stubCallMessages(t *testing.T, fn func(context.Context, string, string, string, map[string]string) (string, error)) func() {
	t.Helper()
	prev := callMessages
	callMessages = func(ctx context.Context, apiKey, model, userMsg string, files map[string]string) (string, error) {
		return fn(ctx, apiKey, model, userMsg, files)
	}
	return func() {
		callMessages = prev
	}
}

func withFiles(m map[string]string) func() {
	prev := files
	files = m
	return func() { files = prev }
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	defer withFiles(map[string]string{"readme.md": "content"})()

	req := httptest.NewRequest(http.MethodGet, "/api/chat", nil)
	rr := httptest.NewRecorder()

	Handler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandler_InvalidJSON(t *testing.T) {
	defer withFiles(map[string]string{"readme.md": "content"})()

	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	Handler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestHandler_EasterEgg(t *testing.T) {
	defer withFiles(map[string]string{"readme.md": "content"})()
	restore := stubCallMessages(t, func(context.Context, string, string, string, map[string]string) (string, error) {
		t.Fatalf("callMessages should not be invoked when Easter egg triggers")
		return "", nil
	})
	defer restore()

	payload := map[string]string{"message": "Tell me an Easter secret"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	Handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var resp struct {
		Output string `json:"output"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	expected, _ := easter.Message("easter")
	if resp.Output != expected {
		t.Fatalf("output = %q, want %q", resp.Output, expected)
	}
}

func TestHandler_Success(t *testing.T) {
	defer withFiles(map[string]string{"readme.md": "context"})()
	t.Setenv("ANTHROPIC_API_KEY", "test-key")

	var received struct {
		apiKey string
		model  string
		msg    string
		files  map[string]string
	}

	restore := stubCallMessages(t, func(ctx context.Context, apiKey, model, msg string, f map[string]string) (string, error) {
		received = struct {
			apiKey string
			model  string
			msg    string
			files  map[string]string
		}{apiKey: apiKey, model: model, msg: msg, files: f}
		return "ok", nil
	})
	defer restore()

	payload := map[string]string{"message": "Hello"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	Handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	if received.apiKey != "test-key" {
		t.Fatalf("api key = %q, want test-key", received.apiKey)
	}
	if received.model != ai.DefaultModel {
		t.Fatalf("model = %q, want %s", received.model, ai.DefaultModel)
	}
	if want := "Format any responses to the question that have lists in human readable bullet points: Hello"; received.msg != want {
		t.Fatalf("message = %q, want %q", received.msg, want)
	}
	if _, ok := received.files["readme.md"]; !ok {
		t.Fatalf("files map missing readme.md: %#v", received.files)
	}

	var resp struct {
		Output string `json:"output"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Output != "ok" {
		t.Fatalf("output = %q, want ok", resp.Output)
	}
}

func TestHandler_UpstreamError(t *testing.T) {
	defer withFiles(map[string]string{"readme.md": "context"})()
	t.Setenv("ANTHROPIC_API_KEY", "test-key")

	restore := stubCallMessages(t, func(context.Context, string, string, string, map[string]string) (string, error) {
		return "", errors.New("boom")
	})
	defer restore()

	payload := map[string]string{"message": "Hello"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	Handler(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadGateway)
	}
}
