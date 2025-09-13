package models

// ChatRequest is the payload for POST /chat.
// It is exported for reuse across packages and deployments.
type ChatRequest struct {
    Message string `json:"message"`
}

// ChatResponse is the JSON response returned by /chat.
// It is exported for reuse across packages and deployments.
type ChatResponse struct {
    Output string `json:"output"`
}

