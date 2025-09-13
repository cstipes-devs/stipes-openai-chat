package embedded

import (
    "strconv"
    "strings"
    "testing"
)

// TestFilesMap_LoadsAssetsAndFormats verifies that FilesMap returns all
// embedded assets, returns exact contents for text files, and summarizes
// non-text files as a binary note including byte size.
func TestFilesMap_LoadsAssetsAndFormats(t *testing.T) {
    got := FilesMap()

    entries, err := FS.ReadDir("assets")
    if err != nil {
        t.Fatalf("ReadDir assets: %v", err)
    }

    if len(got) != len(entries) {
        t.Fatalf("expected %d assets, got %d", len(entries), len(got))
    }

    for _, e := range entries {
        name := e.Name()
        v, ok := got[name]
        if !ok {
            t.Fatalf("missing entry for %q", name)
        }

        b, err := FS.ReadFile("assets/" + name)
        if err != nil {
            t.Fatalf("ReadFile %q: %v", name, err)
        }

        lower := strings.ToLower(name)
        if strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".txt") {
            if v != string(b) {
                t.Errorf("text content mismatch for %q", name)
            }
            continue
        }

        // Expect binary summary: "[binary file loaded: <name>, <n> bytes]"
        prefix := "[binary file loaded: " + name + ", "
        suffix := " bytes]"
        if !strings.HasPrefix(v, prefix) || !strings.HasSuffix(v, suffix) {
            t.Errorf("binary summary format mismatch for %q: %q", name, v)
            continue
        }
        sizeStr := strings.TrimSuffix(strings.TrimPrefix(v, prefix), suffix)
        n, err := strconv.Atoi(sizeStr)
        if err != nil {
            t.Errorf("invalid size in summary for %q: %q", name, v)
            continue
        }
        if n != len(b) {
            t.Errorf("size mismatch for %q: got %d, want %d", name, n, len(b))
        }
    }
}

