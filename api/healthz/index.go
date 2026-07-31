package handler

import (
	"io"
	"net/http"
)

// Handler implements a Vercel Go Serverless Function for /api/healthz
// See: https://vercel.com/docs/functions/runtimes/go
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"ok":true}`)
}
