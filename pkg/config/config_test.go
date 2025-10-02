package config

import "testing"

func TestGet_ReturnsEnvValue(t *testing.T) {
	t.Setenv("SOME_KEY", "value")
	if got := Get("SOME_KEY", "default"); got != "value" {
		t.Fatalf("Get returned %q, want value", got)
	}
}

func TestGet_ReturnsDefaultWhenEmpty(t *testing.T) {
	t.Setenv("EMPTY_KEY", "")
	if got := Get("EMPTY_KEY", "fallback"); got != "fallback" {
		t.Fatalf("Get returned %q, want fallback", got)
	}
}
