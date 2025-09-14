package handler

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chris-stipes/stipes-openai-chat/internal/models"
	sdk "github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/responses"
	"github.com/openai/openai-go/v2/shared"
)

//go:embed ./*
var FS embed.FS

var files = FilesMap()

type ChatRequest struct {
	Message string `json:"message"`
}

// ChatResponse is the JSON response returned by /chat.
// It is exported for reuse across packages and deployments.
type ChatResponse struct {
	Output string `json:"output"`
}

// FilesMap returns filename -> content (text) or a short note for binaries.
func FilesMap() map[string]string {
	out := make(map[string]string)
	entries, err := FS.ReadDir("assets")
	if err != nil {
		log.Printf("failed to read embedded assets: %v", err)
		return out
	}
	for _, e := range entries {
		name := e.Name()
		b, err := FS.ReadFile("assets/" + name)
		if err != nil {
			log.Printf("warn: cannot read %s: %v", name, err)
			continue
		}
		lower := strings.ToLower(name)
		if strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".txt") {
			out[name] = string(b)
		} else {
			out[name] = fmt.Sprintf("[binary file loaded: %s, %d bytes]", name, len(b))
		}
	}
	return out
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid json: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	out, err := CallResponses(ctx, os.Getenv("OPENAI_API_KEY"), "gpt-4o-mini", req.Message, files)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "openai error: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(models.ChatResponse{Output: out})
}

// CallResponses invokes the OpenAI Responses API via the official SDK and returns output text.
func CallResponses(ctx context.Context, apiKey, model, userMsg string, files map[string]string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY is not set")
	}

	// Build a compact context string from embedded files.
	var b strings.Builder
	b.WriteString("You have access to the following project-embedded files. Use them as authoritative context when relevant.\n\n")
	for name, content := range files {
		// Skip non-text placeholders like "[binary file loaded: ...]".
		if strings.HasPrefix(content, "[binary file loaded:") {
			continue
		}
		b.WriteString("---- FILE: ")
		b.WriteString(name)
		b.WriteString(" ----\n")
		const max = 20_000
		if len(content) > max {
			b.WriteString(content[:max])
			b.WriteString("\n[...truncated...]\n")
		} else {
			b.WriteString(content)
		}
		b.WriteString("\n\n")
	}
	contextBlock := b.String()

	client := sdk.NewClient(option.WithAPIKey(apiKey))

	// Compose input as messages: system instructions, context, then user.
	input := responses.ResponseInputParam{
		responses.ResponseInputItemParamOfMessage(
			"You are a concise, helpful engineering assistant. Prefer facts from the provided embedded files.",
			responses.EasyInputMessageRoleSystem,
		),
		responses.ResponseInputItemParamOfMessage(contextBlock, responses.EasyInputMessageRoleSystem),
		responses.ResponseInputItemParamOfMessage(userMsg, responses.EasyInputMessageRoleUser),
	}

	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: shared.ResponsesModel(model),
		Input: responses.ResponseNewParamsInputUnion{OfInputItemList: input},
		Text:  responses.ResponseTextConfigParam{},
	})
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "(no response)", nil
	}
	out := resp.OutputText()
	if out == "" {
		return "(no text output received)", nil
	}
	return out, nil
}
