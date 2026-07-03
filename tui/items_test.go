package main

import (
	"testing"

	"scriptstui/internal/meta"
)

func TestScriptItemFields(t *testing.T) {
	item := scriptItem{m: meta.ScriptMeta{
		Name:        "portkill",
		Description: "kill a port",
		Category:    "net",
	}}
	if item.Title() != "portkill" {
		t.Fatalf("Title = %q", item.Title())
	}
	if item.Description() != "kill a port" {
		t.Fatalf("Description = %q", item.Description())
	}
	if item.FilterValue() != "portkill net kill a port" {
		t.Fatalf("FilterValue = %q", item.FilterValue())
	}
}

func TestScriptItemUndocumented(t *testing.T) {
	item := scriptItem{m: meta.ScriptMeta{Name: "raw.py", Category: "dev"}}
	if item.Description() != "(undocumented)" {
		t.Fatalf("Description = %q", item.Description())
	}
}

func TestNewItemsCount(t *testing.T) {
	items := newItems([]meta.ScriptMeta{{Name: "a"}, {Name: "b"}})
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}
