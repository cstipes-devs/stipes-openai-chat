# stipes-openai-chat

A tiny Go service that:
- Embeds local context files using `go:embed`
- Loads them into memory on startup
- Exposes `POST /chat` which calls the **Claude Messages API** and includes the embedded file contents as context

> Files included:
> - `pkg/embedded/assets/stipes.md`
> - `pkg/embedded/assets/resume.md`

## Prereqs
- Go 1.24+
- An Anthropic API key

## Run
```bash
export ANTHROPIC_API_KEY=sk-ant-...  # required
export ANTHROPIC_MODEL=claude-opus-5-5  # optional (defaults to claude-opus-5-5)
export ADDR=:8080                      # optional

# from repo root
go run ./cmd/server
```

## Endpoints

### Health
```
GET /healthz
```
Returns `200` `application/json` with body `{"ok":true}`. The method is not enforced; any method returns the same response.

### List loaded files
```
GET /files
```
Returns `200` `application/json` with body `{"files":{"<name>":"<content>"}}`. The method is not enforced; any method returns the same response.

### Chat
```
POST /chat
Content-Type: application/json

{
  "message": "Write a short bio of Chris"
}
```
Response:
```json
{
  "output": "…model response…"
}
```

Error responses (note these are **not** JSON):
- `405` — non-POST method; empty body.
- `400` — malformed JSON; `text/plain` body `invalid json: <details>`.
- `502` — upstream Claude API failure; `text/plain` body `upstream error: <details>`.

Clients should check the status code before calling `res.json()`.

## Tests
```bash
go test ./...   # unit tests; no network, safe for CI
```

An integration test calls the real Claude API through the `/chat` handler. It is behind the
`integration` build tag, so `go test ./...` never runs it, and it skips when `ANTHROPIC_API_KEY`
is unset. Run it locally to verify the end-to-end setup (each run costs a few cents):
```bash
ANTHROPIC_API_KEY=sk-ant-... go test -tags=integration -count=1 -v ./internal/handlers/ -run Integration
```

## Notes
- Binary files are summarized with a short note; prefer `.md` or `.txt` for contextual content.
- `handleChat` truncates very long files to keep request size reasonable.

## Deploy: Vercel

You can deploy this repo to Vercel in two ways. Choose one approach per project (Docker vs. Serverless) — if a `Dockerfile` is present, Vercel treats the project as a Container deploy and will not build Serverless Functions.

### Serverless Functions
- Layout: Go functions in `api/` with `func Handler(w http.ResponseWriter, r *http.Request)`.
- Endpoints (built by Vercel):
  - `GET /api/healthz`
  - `GET /api/files`
  - `POST /api/chat`
- Routing: `vercel.json` maps `/api/healthz` and `/api/files` to their functions; any other path falls through to the chat function.
- Behavior difference: the Vercel chat function prepends a formatting instruction to `message` before calling Claude, so outputs can differ from the standalone server for the same input.
- Env vars (set in Vercel Project Settings → Environment Variables):
  - `ANTHROPIC_API_KEY` (required)
  - `ANTHROPIC_MODEL` (optional, default `claude-opus-5-5`)
- Example:
  ```bash
  curl -s http://localhost:3000/api/healthz
  curl -s http://localhost:3000/api/files
  curl -s -X POST http://localhost:3000/api/chat \
    -H 'Content-Type: application/json' \
    -d '{"message":"Write a short bio of Chris"}'
  ```

### Container (Docker)
- Uses root `Dockerfile` (not currently committed — add one to use this path). The app listens on `:$PORT` if set (Vercel sets this), otherwise `ADDR`.
- Endpoints:
  - `GET /healthz`
  - `GET /files`
  - `POST /chat`
- Env vars:
  - `ANTHROPIC_API_KEY` (required)
  - `ANTHROPIC_MODEL` (optional)
  - `ADDR` (optional; ignored if `PORT` is set by the platform)
- Local test:
  ```bash
  docker build -t stipes-openai-chat .
  docker run -e ANTHROPIC_API_KEY=sk-ant-... -p 8080:8080 stipes-openai-chat
  curl -s localhost:8080/healthz
  ```
