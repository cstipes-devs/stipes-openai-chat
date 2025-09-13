package main

import (
    "log"
    "net/http"
    "os"

    embedded "github.com/chris-stipes/stipes-openai-chat"
    "github.com/chris-stipes/stipes-openai-chat/internal/config"
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

    addr := config.Get("ADDR", ":8080")
    log.Printf("listening on %s", addr)
    if err := http.ListenAndServe(addr, middleware.LogReq(mux)); err != nil {
        log.Fatal(err)
    }
}
