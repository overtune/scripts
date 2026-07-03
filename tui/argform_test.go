package main

import (
	"testing"

	"scriptstui/internal/meta"
)

func TestMissingRequired(t *testing.T) {
	args := []meta.Arg{
		{Name: "port", Required: true},
		{Name: "signal", Required: false},
	}
	missing := missingRequired(args, []string{"", "TERM"})
	if len(missing) != 1 || missing[0] != "port" {
		t.Fatalf("expected [port], got %v", missing)
	}

	none := missingRequired(args, []string{"8080", ""})
	if len(none) != 0 {
		t.Fatalf("expected no missing, got %v", none)
	}
}

func TestArgFormValues(t *testing.T) {
	f := newArgForm([]meta.Arg{{Name: "port"}, {Name: "signal"}})
	if len(f.values()) != 2 {
		t.Fatalf("expected 2 values, got %d", len(f.values()))
	}
}
