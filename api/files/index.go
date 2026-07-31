package handler

import (
	"encoding/json"
	"net/http"

	embedded "github.com/chris-stipes/stipes-openai-chat/pkg/embedded"
)

var files = embedded.FilesMap()

// Handler implements a Vercel Go Serverless Function for /api/files
// See: https://vercel.com/docs/functions/runtimes/go
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"files": files})
}
