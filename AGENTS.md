# Repository Guidelines

## Project Structure & Module Organization
- `main.go`: Single-binary HTTP service (handlers, Claude call, embed init).
- `assets/`: Embedded context files loaded at startup.
- `.env.example`: Documented env vars (`ANTHROPIC_API_KEY`, `ANTHROPIC_MODEL`, `ADDR`).
- `README.md`: Usage and API reference. Go module is defined in `go.mod`.

## Build, Run, and Test
- Run locally: `go run ./cmd/server` (requires `ANTHROPIC_API_KEY`).
- Build binary: `go build -o bin/server ./cmd/server`
- Unit tests: `go test ./...`
- Coverage: `go test ./... -cover`
- Integration (real Claude API, local only; excluded from CI by build tag): `ANTHROPIC_API_KEY=... go test -tags=integration -count=1 -v ./internal/handlers/ -run Integration`
- Health check: `curl localhost:8080/healthz`
- Chat example:
  ```bash
  curl -s -X POST localhost:8080/chat \
    -H 'Content-Type: application/json' \
    -d '{"message":"Write a short bio of Chris"}'
  ```

## Coding Style & Naming Conventions
- Use `go fmt ./...` before pushing; keep imports organized.
- Package and file names: lowercase with underscores sparingly; prefer short, clear names.
- Exported identifiers: `MixedCaps` with doc comments when non-obvious.
- Error handling: wrap with context (`fmt.Errorf("...: %w", err)`). Prefer early returns.
- HTTP handlers: keep small and pure; move helpers to top-level functions.

## Testing Guidelines
- Framework: standard `testing` package; name files `*_test.go`.
- Test names: `TestFunction_Scenario` and table-driven where useful.
- Cover edge cases (empty input, timeouts, Claude API error paths). Use small fakes over network calls.
- Run `go test ./...` locally; add tests when changing handlers or request building.

## Commit & Pull Request Guidelines
- Commits: concise, imperative subject (max ~72 chars), body explaining intent and impact.
  - Example: `feat(chat): truncate long files for request size`
- PRs: include summary, screenshots or curl output when API behavior changes, and link related issues.
- Keep diffs focused; update `README.md` and `.env.example` when endpoints or env vars change.

## Security & Configuration Tips
- Do not commit secrets. Load config via env or a local `.env` (untracked).
- Required: `ANTHROPIC_API_KEY`. Optional: `ANTHROPIC_MODEL` (defaults `claude-opus-5-5`), `ADDR` (defaults `:8080`).
- Large/binary assets are embedded but not parsed; keep them minimal to control request size.
