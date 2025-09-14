package models

// ChatRequest is the payload for POST /chat.
type ChatRequest struct {
    Message string `json:"message"`
}

// ChatResponse is the JSON response returned by /chat.
type ChatResponse struct {
    Output string `json:"output"`
}

