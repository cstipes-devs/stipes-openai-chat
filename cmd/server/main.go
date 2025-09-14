package main

import (
	"log"
	"net/http"
	"os"

	"github.com/chris-stipes/stipes-openai-chat/pkg/config"
	embedded "github.com/chris-stipes/stipes-openai-chat/pkg/embedded"
	"github.com/chris-stipes/stipes-openai-chat/internal/handlers"
	"github.com/chris-stipes/stipes-openai-chat/internal/middleware"
)

func main() {
	mux := http.NewServeMux()
	handlers.Register(mux, handlers.Deps{
		Files:  embedded.FilesMap(),
		APIKey: os.Getenv("OPENAI_API_KEY"),
		Model:  config.Get("OPENAI_MODEL", "gpt-4o-mini"),
	})

	// Prefer the platform-provided PORT if present (e.g., Vercel),
	// otherwise fall back to ADDR or :8080.
	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	} else {
		addr = config.Get("ADDR", ":8080")
	}
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, middleware.LogReq(mux)); err != nil {
		log.Fatal(err)
	}
}
