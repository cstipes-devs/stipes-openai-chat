//go:build integration

// Integration tests call the real Claude API. They are excluded from
// `go test ./...` by the build tag and skip when ANTHROPIC_API_KEY is unset.
//
//	ANTHROPIC_API_KEY=sk-ant-... go test -tags=integration -count=1 -v ./internal/handlers/ -run Integration
package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/chris-stipes/stipes-openai-chat/pkg/claude"
	"github.com/chris-stipes/stipes-openai-chat/pkg/config"
	embedded "github.com/chris-stipes/stipes-openai-chat/pkg/embedded"
	"github.com/chris-stipes/stipes-openai-chat/pkg/models"
)

func newIntegrationServer(t *testing.T, apiKey string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	Register(mux, Deps{
		Files:  embedded.FilesMap(),
		APIKey: apiKey,
		Model:  config.Get("ANTHROPIC_MODEL", claude.DefaultModel),
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func postChat(t *testing.T, srv *httptest.Server, message string) (int, string) {
	t.Helper()
	body, _ := json.Marshal(models.ChatRequest{Message: message})
	resp, err := http.Post(srv.URL+"/chat", "application/json", strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("POST /chat: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(raw)
}

func TestChatIntegration_RealClaude(t *testing.T) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}
	srv := newIntegrationServer(t, key)

	status, raw := postChat(t, srv, "Write a short bio of Chris")
	if status != http.StatusOK {
		t.Fatalf("status = %d, body = %q", status, raw)
	}
	var resp models.ChatResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("decode response %q: %v", raw, err)
	}
	t.Logf("model output:\n%s", resp.Output)

	if strings.HasPrefix(resp.Output, "(no ") {
		t.Fatalf("expected model text, got placeholder %q", resp.Output)
	}
	// The question asks about Chris, so a grounded answer names him; wording is otherwise free.
	lower := strings.ToLower(resp.Output)
	if !strings.Contains(lower, "chris") && !strings.Contains(lower, "stipes") {
		t.Fatalf("expected answer to mention Chris or Stipes, got %q", resp.Output)
	}
}

func TestChatIntegration_BadKeyReturns502(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}
	srv := newIntegrationServer(t, "sk-ant-invalid-key-for-integration-test")

	status, raw := postChat(t, srv, "hello")
	if status != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d; body = %q", status, http.StatusBadGateway, raw)
	}
	if !strings.Contains(raw, "upstream error:") {
		t.Fatalf("expected upstream error prefix, got %q", raw)
	}
}
