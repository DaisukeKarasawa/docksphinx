package main

import "testing"

func TestFormatFloat1OrNA(t *testing.T) {
	if got := formatFloat1OrNA(12.34, true); got != "12.3" {
		t.Fatalf("expected 12.3, got %s", got)
	}
	if got := formatFloat1OrNA(12.34, false); got != "N/A" {
		t.Fatalf("expected N/A, got %s", got)
	}
}

func TestFormatInt64OrNA(t *testing.T) {
	if got := formatInt64OrNA(123, true); got != "123" {
		t.Fatalf("expected 123, got %s", got)
	}
	if got := formatInt64OrNA(123, false); got != "N/A" {
		t.Fatalf("expected N/A, got %s", got)
	}
}
