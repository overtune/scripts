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
	if item.Title() != "net / portkill" {
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

func TestNewItemsSortedByCategoryThenName(t *testing.T) {
	input := []meta.ScriptMeta{
		{Name: "zeta", Category: "net"},
		{Name: "alpha", Category: "dev"},
		{Name: "beta", Category: "net"},
	}
	items := newItems(input)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	var got []meta.ScriptMeta
	for _, it := range items {
		si, ok := it.(scriptItem)
		if !ok {
			t.Fatalf("item is not a scriptItem: %#v", it)
		}
		got = append(got, si.m)
	}

	want := []struct {
		category, name string
	}{
		{"dev", "alpha"},
		{"net", "beta"},
		{"net", "zeta"},
	}
	for i, w := range want {
		if got[i].Category != w.category || got[i].Name != w.name {
			t.Fatalf("item %d = (%s, %s), want (%s, %s)", i, got[i].Category, got[i].Name, w.category, w.name)
		}
	}
}
