package claude

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	sdk "github.com/anthropics/anthropic-sdk-go"
)

func restoreSendMessage(t *testing.T, fn func(context.Context, string, sdk.BetaMessageNewParams) (*sdk.BetaMessage, error)) func() {
	t.Helper()
	prev := sendMessage
	sendMessage = fn
	return func() { sendMessage = prev }
}

// fakeMessage builds a response the same way the SDK does: by unmarshaling API JSON.
func fakeMessage(t *testing.T, raw string) *sdk.BetaMessage {
	t.Helper()
	var msg sdk.BetaMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal fake message: %v", err)
	}
	return &msg
}

func TestCallMessages_RequiresAPIKey(t *testing.T) {
	if _, err := CallMessages(context.Background(), "", "model", "msg", nil); err == nil {
		t.Fatalf("expected error when API key missing")
	}
}

func TestCallMessages_ReturnsOutputText(t *testing.T) {
	var captured sdk.BetaMessageNewParams
	restore := restoreSendMessage(t, func(ctx context.Context, apiKey string, params sdk.BetaMessageNewParams) (*sdk.BetaMessage, error) {
		if apiKey != "key" {
			return nil, errors.New("unexpected api key")
		}
		captured = params
		return fakeMessage(t, `{"stop_reason":"end_turn","content":[
			{"type":"thinking","thinking":"","signature":"sig"},
			{"type":"text","text":"ans"},
			{"type":"text","text":"wer"}]}`), nil
	})
	defer restore()

	files := map[string]string{
		"notes.md": "hello",
		"huge.md":  strings.Repeat("x", 20_050),
		"bin.dat":  "[binary file loaded: bin.dat, 3 bytes]",
	}

	out, err := CallMessages(context.Background(), "key", "claude-test", "question", files)
	if err != nil {
		t.Fatalf("CallMessages returned error: %v", err)
	}
	if out != "answer" {
		t.Fatalf("output = %q, want answer", out)
	}
	if captured.Model != "claude-test" {
		t.Fatalf("model = %q, want claude-test", captured.Model)
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
	if !strings.Contains(marshaled, `"cache_control"`) {
		t.Fatalf("expected cache_control on context block, got %s", marshaled)
	}
}

func TestCallMessages_ContextIsDeterministic(t *testing.T) {
	var systems []string
	restore := restoreSendMessage(t, func(_ context.Context, _ string, params sdk.BetaMessageNewParams) (*sdk.BetaMessage, error) {
		raw, _ := json.Marshal(params.System)
		systems = append(systems, string(raw))
		return fakeMessage(t, `{"stop_reason":"end_turn","content":[{"type":"text","text":"ok"}]}`), nil
	})
	defer restore()

	files := map[string]string{"a.md": "A", "b.md": "B", "c.md": "C", "d.md": "D", "e.md": "E"}
	for i := 0; i < 10; i++ {
		if _, err := CallMessages(context.Background(), "key", "model", "msg", files); err != nil {
			t.Fatalf("CallMessages returned error: %v", err)
		}
	}
	for i, s := range systems {
		if s != systems[0] {
			t.Fatalf("system prompt differs on call %d; caching would miss", i)
		}
	}
	if !strings.Contains(systems[0], "a.md ----\\nA\\n\\n---- FILE: b.md") {
		t.Fatalf("expected files in sorted order, got %s", systems[0])
	}
}

func TestCallMessages_StopReasonErrors(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"refusal", `{"stop_reason":"refusal","stop_details":{"type":"refusal","category":"cyber","explanation":"nope"},"content":[]}`, "refused"},
		{"max_tokens", `{"stop_reason":"max_tokens","content":[{"type":"text","text":"partial"}]}`, "max_tokens"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := restoreSendMessage(t, func(context.Context, string, sdk.BetaMessageNewParams) (*sdk.BetaMessage, error) {
				return fakeMessage(t, tt.raw), nil
			})
			defer restore()

			_, err := CallMessages(context.Background(), "key", "model", "msg", nil)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestCallMessages_NilResponse(t *testing.T) {
	restore := restoreSendMessage(t, func(context.Context, string, sdk.BetaMessageNewParams) (*sdk.BetaMessage, error) {
		return nil, nil
	})
	defer restore()

	out, err := CallMessages(context.Background(), "key", "model", "msg", map[string]string{"a.md": "text"})
	if err != nil {
		t.Fatalf("CallMessages returned error: %v", err)
	}
	if out != "(no response)" {
		t.Fatalf("output = %q, want (no response)", out)
	}
}

func TestCallMessages_EmptyOutput(t *testing.T) {
	restore := restoreSendMessage(t, func(context.Context, string, sdk.BetaMessageNewParams) (*sdk.BetaMessage, error) {
		return fakeMessage(t, `{"stop_reason":"end_turn","content":[{"type":"text","text":""}]}`), nil
	})
	defer restore()

	out, err := CallMessages(context.Background(), "key", "model", "msg", map[string]string{"a.md": "text"})
	if err != nil {
		t.Fatalf("CallMessages returned error: %v", err)
	}
	if out != "(no text output received)" {
		t.Fatalf("output = %q, want (no text output received)", out)
	}
}

func TestCallMessages_PropagatesAPIError(t *testing.T) {
	boom := errors.New("boom")
	restore := restoreSendMessage(t, func(context.Context, string, sdk.BetaMessageNewParams) (*sdk.BetaMessage, error) {
		return nil, boom
	})
	defer restore()

	_, err := CallMessages(context.Background(), "key", "model", "msg", map[string]string{"a.md": "text"})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want wrapping boom", err)
	}
}
