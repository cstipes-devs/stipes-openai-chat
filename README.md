# stipes-openai-chat

A tiny Go service that:
- Embeds local context files using `go:embed`
- Loads them into memory on startup
- Exposes `POST /chat` which calls the **OpenAI Responses API** and includes the embedded file contents as context

> Files included:
> - `internal/embedded/assets/stipes.md`
> - `internal/embedded/assets/resume.md`

## Prereqs
- Go 1.22+
- An OpenAI API key

## Run
```bash
export OPENAI_API_KEY=sk-...        # required
export OPENAI_MODEL=gpt-4o-mini     # optional (defaults to gpt-4o-mini)
export ADDR=:8080                   # optional

# from repo root
go run ./cmd/server
```

## Endpoints

### Health
```
GET /healthz
```

### List loaded files
```
GET /files
```

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
- Env vars (set in Vercel Project Settings → Environment Variables):
  - `OPENAI_API_KEY` (required)
  - `OPENAI_MODEL` (optional, default `gpt-4o-mini`)
- Example:
  ```bash
  curl -s http://localhost:3000/api/healthz
  curl -s http://localhost:3000/api/files
  curl -s -X POST http://localhost:3000/api/chat \
    -H 'Content-Type: application/json' \
    -d '{"message":"Write a short bio of Chris"}'
  ```

### Container (Docker)
- Uses root `Dockerfile`. The app listens on `:$PORT` if set (Vercel sets this), otherwise `ADDR`.
- Endpoints:
  - `GET /healthz`
  - `GET /files`
  - `POST /chat`
- Env vars:
  - `OPENAI_API_KEY` (required)
  - `OPENAI_MODEL` (optional)
  - `ADDR` (optional; ignored if `PORT` is set by the platform)
- Local test:
  ```bash
  docker build -t stipes-openai-chat .
  docker run -e OPENAI_API_KEY=sk-... -p 8080:8080 stipes-openai-chat
  curl -s localhost:8080/healthz
  ```
