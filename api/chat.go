package handler

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"

    embedded "github.com/chris-stipes/stipes-openai-chat/pkg/embedded"
    "github.com/chris-stipes/stipes-openai-chat/pkg/config"
    "github.com/chris-stipes/stipes-openai-chat/pkg/models"
    ai "github.com/chris-stipes/stipes-openai-chat/pkg/openai"
)

var files = embedded.FilesMap()

// Handler implements a Vercel Go Serverless Function for /api/chat
// See: https://vercel.com/docs/functions/runtimes/go
func Handler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var req models.ChatRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprintf(w, "invalid json: %v", err)
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
    defer cancel()

    out, err := ai.CallResponses(ctx, os.Getenv("OPENAI_API_KEY"), config.Get("OPENAI_MODEL", "gpt-4o-mini"), req.Message, files)
    if err != nil {
        w.WriteHeader(http.StatusBadGateway)
        fmt.Fprintf(w, "openai error: %v", err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(models.ChatResponse{Output: out})
}
