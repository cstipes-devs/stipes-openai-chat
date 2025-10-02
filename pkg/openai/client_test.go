package openai

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/openai/openai-go/v2/responses"
)

func restoreSendResponse(t *testing.T, fn func(context.Context, string, responses.ResponseNewParams) (*responses.Response, error)) func() {
	t.Helper()
	prev := sendResponse
	sendResponse = fn
	return func() { sendResponse = prev }
}

func TestCallResponses_RequiresAPIKey(t *testing.T) {
	if _, err := CallResponses(context.Background(), "", "model", "msg", nil); err == nil {
		t.Fatalf("expected error when API key missing")
	}
}

func TestCallResponses_ReturnsOutputText(t *testing.T) {
	var captured responses.ResponseNewParams
	restore := restoreSendResponse(t, func(ctx context.Context, apiKey string, params responses.ResponseNewParams) (*responses.Response, error) {
		if apiKey != "key" {
			return nil, errors.New("unexpected api key")
		}
		captured = params
		return &responses.Response{
			Output: []responses.ResponseOutputItemUnion{
				{Content: []responses.ResponseOutputMessageContentUnion{{Type: "output_text", Text: "answer"}}},
			},
		}, nil
	})
	defer restore()

	files := map[string]string{
		"notes.md": "hello",
		"huge.md":  strings.Repeat("x", 20_050),
		"bin.dat":  "[binary file loaded: bin.dat, 3 bytes]",
	}

	out, err := CallResponses(context.Background(), "key", "gpt-4o", "question", files)
	if err != nil {
		t.Fatalf("CallResponses returned error: %v", err)
	}
	if out != "answer" {
		t.Fatalf("output = %q, want answer", out)
	}
	if captured.Model != "gpt-4o" {
		t.Fatalf("model = %q, want gpt-4o", captured.Model)
	}

	raw, err := json.Marshal(captured)
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}
	marshaled := string(raw)
	if !strings.Contains(marshaled, "notes.md") {
		t.Fatalf("expected params to include notes.md, got %s", marshaled)
	}
	if !strings.Contains(marshaled, "[...truncated...]") {
		t.Fatalf("expected truncated marker in params, got %s", marshaled)
	}
	if strings.Contains(marshaled, "bin.dat, 3 bytes]") {
		t.Fatalf("binary placeholder should be skipped from context: %s", marshaled)
	}
}

func TestCallResponses_NilResponse(t *testing.T) {
	restore := restoreSendResponse(t, func(context.Context, string, responses.ResponseNewParams) (*responses.Response, error) {
		return nil, nil
	})
	defer restore()

	out, err := CallResponses(context.Background(), "key", "model", "msg", map[string]string{"a.md": "text"})
	if err != nil {
		t.Fatalf("CallResponses returned error: %v", err)
	}
	if out != "(no response)" {
		t.Fatalf("output = %q, want (no response)", out)
	}
}

func TestCallResponses_EmptyOutput(t *testing.T) {
	restore := restoreSendResponse(t, func(context.Context, string, responses.ResponseNewParams) (*responses.Response, error) {
		return &responses.Response{Output: []responses.ResponseOutputItemUnion{
			{Content: []responses.ResponseOutputMessageContentUnion{{Type: "output_text", Text: ""}}},
		}}, nil
	})
	defer restore()

	out, err := CallResponses(context.Background(), "key", "model", "msg", map[string]string{"a.md": "text"})
	if err != nil {
		t.Fatalf("CallResponses returned error: %v", err)
	}
	if out != "(no text output received)" {
		t.Fatalf("output = %q, want (no text output received)", out)
	}
}

func TestCallResponses_PropagatesAPIError(t *testing.T) {
	restore := restoreSendResponse(t, func(context.Context, string, responses.ResponseNewParams) (*responses.Response, error) {
		return nil, errors.New("boom")
	})
	defer restore()

	if _, err := CallResponses(context.Background(), "key", "model", "msg", map[string]string{"a.md": "text"}); err == nil {
		t.Fatalf("expected propagated error")
	}
}
