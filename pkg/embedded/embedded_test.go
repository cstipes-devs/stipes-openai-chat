package embedded

import (
	"strings"
	"testing"
)

func TestFilesMap_LoadsAssets(t *testing.T) {
	files := FilesMap()

	text, ok := files["resume.md"]
	if !ok {
		t.Fatalf("resume.md missing from files map")
	}
	if text == "" {
		t.Fatalf("resume.md content should not be empty")
	}

	bin, ok := files["sample.bin"]
	if !ok {
		t.Fatalf("sample.bin missing from files map")
	}
	wantPrefix := "[binary file loaded: sample.bin"
	if !strings.HasPrefix(bin, wantPrefix) {
		t.Fatalf("binary placeholder = %q, want prefix %q", bin, wantPrefix)
	}
}
