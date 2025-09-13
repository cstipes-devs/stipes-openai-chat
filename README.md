# stipes-openai-chat

A tiny Go service that:
- Embeds local context files using `go:embed`
- Loads them into memory on startup
- Exposes `POST /chat` which calls the **OpenAI Responses API** and includes the embedded file contents as context

> Files included:
> - `assets/stipes.md`
> - `assets/Christopher_Stipes_Resume_20250911.pdf`

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
- PDF is loaded but not parsed; the service passes a note with byte length so the model knows it exists.
- `handleChat` truncates very long files to keep request size reasonable.
