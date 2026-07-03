package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"scriptstui/internal/meta"
)

// scriptItem adapts a ScriptMeta to the bubbles list.DefaultItem interface.
type scriptItem struct {
	m meta.ScriptMeta
}

func (i scriptItem) Title() string { return i.m.Name }

func (i scriptItem) Description() string {
	if i.m.Description == "" {
		return "(undocumented)"
	}
	return i.m.Description
}

func (i scriptItem) FilterValue() string {
	return strings.TrimSpace(i.m.Name + " " + i.m.Category + " " + i.m.Description)
}

func newItems(metas []meta.ScriptMeta) []list.Item {
	items := make([]list.Item, 0, len(metas))
	for _, m := range metas {
		items = append(items, scriptItem{m: m})
	}
	return items
}
