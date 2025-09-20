package openai

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/responses"
)

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

	// Compose input as messages: system instructions, context, then user question.
	input := responses.ResponseInputParam{
		// High-level style instruction
		responses.ResponseInputItemParamOfMessage(
			"You are a concise, helpful engineering assistant. Always rewrite retrieved content into clear, polished, human-readable sentences (not raw lists or fragments).",
			responses.EasyInputMessageRoleSystem,
		),
		// Context block from embedded files
		responses.ResponseInputItemParamOfMessage(contextBlock, responses.EasyInputMessageRoleSystem),
		// Wrap the user’s question with explicit formatting instructions
		responses.ResponseInputItemParamOfMessage(
			fmt.Sprintf("Answer the following in one or two natural sentences: %s", userMsg),
			responses.EasyInputMessageRoleUser,
		),
	}

	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: model,
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
