package embedded

import (
    "embed"
    "fmt"
    "log"
    "strings"
)

//go:embed assets/*
var FS embed.FS

// FilesMap returns filename -> content (text) or a short note for binaries.
func FilesMap() map[string]string {
    out := make(map[string]string)
    entries, err := FS.ReadDir("assets")
    if err != nil {
        log.Printf("failed to read embedded assets: %v", err)
        return out
    }
    for _, e := range entries {
        name := e.Name()
        b, err := FS.ReadFile("assets/" + name)
        if err != nil {
            log.Printf("warn: cannot read %s: %v", name, err)
            continue
        }
        lower := strings.ToLower(name)
        if strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".txt") {
            out[name] = string(b)
        } else {
            out[name] = fmt.Sprintf("[binary file loaded: %s, %d bytes]", name, len(b))
        }
    }
    return out
}

