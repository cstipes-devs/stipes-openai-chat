package handlers

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/chris-stipes/stipes-openai-chat/pkg/models"
    ai "github.com/chris-stipes/stipes-openai-chat/pkg/openai"
)

type Deps struct {
    Files map[string]string
    APIKey string
    Model  string
}

// Register wires up all HTTP routes on the provided mux.
func Register(mux *http.ServeMux, d Deps) {
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        io.WriteString(w, `{"ok":true}`)
    })

    mux.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        m := map[string]any{"files": d.Files}
        json.NewEncoder(w).Encode(m)
    })

    mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
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

        out, err := ai.CallResponses(ctx, d.APIKey, d.Model, req.Message, d.Files)
        if err != nil {
            w.WriteHeader(http.StatusBadGateway)
            fmt.Fprintf(w, "openai error: %v", err)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(models.ChatResponse{Output: out})
    })
}
