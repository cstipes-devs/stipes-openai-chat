# How This Chat Bot Works

A short, blog-style tour of the minimal Go service that wraps OpenAI’s Responses API with a clean HTTP interface and a small bit of embedded context.

## TL;DR

- Exposes `POST /chat` backed by OpenAI Responses API.
- Embeds a couple of Markdown files into the binary and sends them as context.
- Ships as a single binary (`go run ./cmd/server`) or a Vercel serverless function (`/api/chat`).

## Architecture

```mermaid
flowchart TB
  A[Client \n curl/Browser] -->|JSON { message }| B[Go HTTP Server \n net/http + mux]
  subgraph S[Service]
    B --> C[Handlers \n /healthz, /files, /chat]
    E[Embedded FS \n pkg/embedded] --> C
    C --> D[OpenAI Client \n pkg/openai]
  end
  D -->|SDK call| O[(OpenAI Responses API)]
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
3. The service calls `pkg/openai.CallResponses` with the API key, model, the user message, and the context.
4. The OpenAI SDK executes a Responses API request with three message blocks:
   - A short system style instruction (concise, helpful tone).
   - A system context block containing embedded file contents.
   - The user message (wrapped with a brief instruction to keep answers short).
5. The output text is returned as `{ "output": "..." }`.
6. Errors map to HTTP status codes:
   - 400 for invalid JSON.
   - 405 for wrong method.
   - 502 for downstream OpenAI errors (e.g., missing/invalid API key).

## Key Components

- `cmd/server/main.go`: Wires routes, loads env, starts the HTTP server.
- `internal/handlers`: Small, focused handlers for `/healthz`, `/files`, and `/chat`.
- `internal/middleware/logging.go`: Logs method, path, and latency.
- `pkg/embedded`: Embeds Markdown assets and exposes `FilesMap()` for handlers.
- `pkg/openai`: Thin wrapper over the official `openai-go` SDK (`Responses` API):
  - Assembles messages (style system prompt, context, user question).
  - Calls `Responses.New(...)` and extracts `resp.OutputText()`.
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

- `OPENAI_API_KEY` (required): API key for OpenAI.
- `OPENAI_MODEL` (optional): Defaults to `gpt-4o-mini`.
- `ADDR` (optional): Server listen address, defaults to `:8080` (or `PORT` if provided by the platform).

## Local Dev & Tests

- Run: `go run ./cmd/server`
- Health: `curl localhost:8080/healthz`
- Chat: see example curl above
- Tests: `go test ./...`

## Notes on Embedded Context

- Markdown files in `pkg/embedded/assets/` are bundled into the binary.
- Large files are truncated to avoid large request payloads.
- Non-text (binary) files are skipped.

## Extending the Bot

- Add new context files: place Markdown in `pkg/embedded/assets/`.
- Tune the tone or format: adjust system/user messages in `pkg/openai/client.go`.
- Swap models: change `OPENAI_MODEL` or expose it per-request.
- Serverless friendly: `api/chat.go` works on Vercel using the same packages.

## Security Considerations

- Do not commit secrets; load `OPENAI_API_KEY` from env or a local untracked `.env`.
- Keep embedded assets minimal to control request size and latency.

