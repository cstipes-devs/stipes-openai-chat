package easter

import "testing"

func TestMessage_FindsEasterCaseInsensitive(t *testing.T) {
	got, ok := Message("Tell me an EaStEr secret")
	if !ok {
		t.Fatalf("expected ok to be true")
	}
	if got != asciiEgg {
		t.Fatalf("message = %q, want %q", got, asciiEgg)
	}
}

func TestMessage_NoMatch(t *testing.T) {
	if _, ok := Message("nothing to see here"); ok {
		t.Fatalf("expected ok to be false")
	}
}
