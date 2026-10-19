package claude

import (
	"context"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// DefaultModel is used when ANTHROPIC_MODEL is unset.
const DefaultModel = "claude-opus-5-5"

var sendMessage = func(ctx context.Context, apiKey string, params sdk.BetaMessageNewParams) (*sdk.BetaMessage, error) {
	client := sdk.NewClient(option.WithAPIKey(apiKey))
	return client.Beta.Messages.New(ctx, params)
}

// CallMessages invokes the Claude Messages API via the official SDK and returns output text.
func CallMessages(ctx context.Context, apiKey, model, userMsg string, files map[string]string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY is not set")
	}

	// Sort file names so the context block is byte-identical across requests;
	// prompt caching only hits on an exact prefix match.
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	// Build a compact context string from embedded files.
	var b strings.Builder
	b.WriteString("You have access to the following project-embedded files. Use them as authoritative context when relevant.\n\n")
	for _, name := range names {
		content := files[name]
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

	resp, err := sendMessage(ctx, apiKey, sdk.BetaMessageNewParams{
		Model: model,
		// Thinking is always on for Opus 5.5 and counts toward this cap, so leave headroom.
		MaxTokens: 16000,
		System: []sdk.BetaTextBlockParam{
			// High-level style instruction
			{Text: "You are a concise, helpful engineering assistant. Always rewrite retrieved content into clear, polished, human-readable sentences (not raw lists or fragments)."},
			// Context block from embedded files; cached since it is identical on every request.
			{Text: contextBlock, CacheControl: sdk.NewBetaCacheControlEphemeralParam()},
		},
		Messages: []sdk.BetaMessageParam{
			// Wrap the user’s question with explicit formatting instructions
			sdk.NewBetaUserMessage(sdk.NewBetaTextBlock(fmt.Sprintf("Answer the following in one or two natural sentences: %s", userMsg))),
		},
		// Short chat answers don't benefit from deeper reasoning.
		OutputConfig: sdk.BetaOutputConfigParam{Effort: sdk.BetaOutputConfigEffortLow},
		// Re-serve safety-classifier refusals on a fallback model chosen by the API.
		Fallbacks: sdk.BetaFallbacksParamOfDefault(),
		Betas:     []sdk.AnthropicBeta{sdk.AnthropicBetaServerSideFallback2026_07_01},
	})
	if err != nil {
		return "", fmt.Errorf("claude messages: %w", err)
	}
	if resp == nil {
		return "(no response)", nil
	}
	switch resp.StopReason {
	case sdk.BetaStopReasonRefusal:
		return "", fmt.Errorf("request refused (category %q): %s", resp.StopDetails.Category, resp.StopDetails.Explanation)
	case sdk.BetaStopReasonMaxTokens:
		return "", fmt.Errorf("response truncated at max_tokens")
	}

	var out strings.Builder
	for _, block := range resp.Content {
		if text, ok := block.AsAny().(sdk.BetaTextBlock); ok {
			out.WriteString(text.Text)
		}
	}
	if out.Len() == 0 {
		return "(no text output received)", nil
	}
	return out.String(), nil
}
