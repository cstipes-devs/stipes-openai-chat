package middleware

import (
    "log"
    "net/http"
    "time"
)

// LogReq logs method, path, and latency for each request.
func LogReq(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
    })
}

