package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	embedded "github.com/chris-stipes/stipes-openai-chat/internal/embedded"
	"github.com/chris-stipes/stipes-openai-chat/internal/models"
	ai "github.com/chris-stipes/stipes-openai-chat/internal/openai"
)

var files = embedded.FilesMap()

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

	out, err := ai.CallResponses(ctx, os.Getenv("OPENAI_API_KEY"), "gpt-4o-mini", req.Message, files)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "openai error: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(models.ChatResponse{Output: out})
}
