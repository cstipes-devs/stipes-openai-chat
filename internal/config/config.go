package config

import "os"

// Get returns env var k or default if empty.
func Get(k, def string) string {
    v := os.Getenv(k)
    if v == "" {
        return def
    }
    return v
}

