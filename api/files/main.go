package main

import (
    "encoding/json"
    "net/http"

    embedded "github.com/chris-stipes/stipes-openai-chat"
)

var files = embedded.FilesMap()

// Handler returns the embedded files map as JSON.
func Handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{"files": files})
}

