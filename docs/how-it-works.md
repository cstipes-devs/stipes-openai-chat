# How This Chat Bot Works

A short, blog-style tour of the minimal Go service that wraps the Claude Messages API with a clean HTTP interface and a small bit of embedded context.

## TL;DR

- Exposes `POST /chat` backed by the Claude Messages API.
- Embeds a couple of Markdown files into the binary and sends them as context.
- Ships as a single binary (`go run ./cmd/server`) or a Vercel serverless function (`/api/chat`).

## Architecture

```mermaid
flowchart TB
  A[Client \n curl/Browser] -->|JSON { message }| B[Go HTTP Server \n net/http + mux]
  subgraph S[Service]
    B --> C[Handlers \n /healthz, /files, /chat]
    E[Embedded FS \n pkg/embedded] --> C
    C --> D[Claude Client \n pkg/claude]
  end
  D -->|SDK call| O[(Claude Messages API)]
  O -->|text output| D --> C -->|JSON { output }| A

  %% Alternate deployment path
  A -.->|POST /api/chat| V[Vercel Function \n api/chat.go]
  V --> E
  V --> D
  V --> A
```

## Request Lifecycle

1. Client sends `POST /chat` with JSON payload: `{ "message": "..." }`.
2. Handler validates method and JSON, then builds a compact context string from embedded files:
   - Skips binary placeholders.
   - Truncates very long files (20k chars) to keep requests small.
3. The service calls `pkg/claude.CallMessages` with the API key, model, the user message, and the context.
4. The Anthropic SDK executes a Messages API request with:
   - A short system style instruction (concise, helpful tone).
   - A system context block containing embedded file contents, in sorted order and marked for prompt caching.
   - The user message (wrapped with a brief instruction to keep answers short).
   - Low effort, since answers are one or two sentences, and server-side refusal fallbacks.
5. The output text is returned as `{ "output": "..." }`.
6. Errors map to HTTP status codes:
   - 400 for invalid JSON.
   - 405 for wrong method.
   - 502 for downstream Claude API errors (e.g., missing/invalid API key).

## Key Components

- `cmd/server/main.go`: Wires routes, loads env, starts the HTTP server.
- `internal/handlers`: Small, focused handlers for `/healthz`, `/files`, and `/chat`.
- `internal/middleware/logging.go`: Logs method, path, and latency.
- `pkg/embedded`: Embeds Markdown assets and exposes `FilesMap()` for handlers.
- `pkg/claude`: Thin wrapper over the official `anthropic-sdk-go` SDK (Messages API):
  - Assembles the system prompt (style instruction, file context) and the user question.
  - Calls `Beta.Messages.New(...)`, maps refusal/`max_tokens` stops to errors, and joins the text blocks.
- `pkg/models`: Shared request/response structs.
- `api/chat.go`: Vercel serverless entrypoint that reuses the same packages (adds a small formatting tweak for list answers).

## Endpoints

- `GET /healthz`: Simple `{ "ok": true }` health check.
- `GET /files`: Returns the embedded files map for transparency.
- `POST /chat`: Accepts `{ "message": "..." }`, returns `{ "output": "..." }`.

Example:

```bash
curl -s -X POST localhost:8080/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"Write a short bio of Chris"}' | jq
```

## Configuration

- `ANTHROPIC_API_KEY` (required): API key for the Claude API.
- `ANTHROPIC_MODEL` (optional): Defaults to `claude-opus-5-5`.
- `ADDR` (optional): Server listen address, defaults to `:8080` (or `PORT` if provided by the platform).

## Local Dev & Tests

- Run: `go run ./cmd/server`
- Health: `curl localhost:8080/healthz`
- Chat: see example curl above
- Tests: `go test ./...`
- Integration (real API, local only): `go test -tags=integration -count=1 -v ./internal/handlers/ -run Integration`

## Notes on Embedded Context

- Markdown files in `pkg/embedded/assets/` are bundled into the binary.
- Large files are truncated to avoid large request payloads.
- Non-text (binary) files are skipped.

## Extending the Bot

- Add new context files: place Markdown in `pkg/embedded/assets/`.
- Tune the tone or format: adjust system/user messages in `pkg/claude/client.go`.
- Swap models: change `ANTHROPIC_MODEL` or expose it per-request.
- Serverless friendly: `api/chat.go` works on Vercel using the same packages.

## Security Considerations

- Do not commit secrets; load `ANTHROPIC_API_KEY` from env or a local untracked `.env`.
- Keep embedded assets minimal to control request size and latency.

