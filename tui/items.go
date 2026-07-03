package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"scriptstui/internal/meta"
)

// scriptItem adapts a ScriptMeta to the bubbles list.DefaultItem interface.
type scriptItem struct {
	m meta.ScriptMeta
}

// Title includes the category first so that, once the list is sorted by
// (Category, Name), items visually cluster into their category groups.
func (i scriptItem) Title() string { return fmt.Sprintf("%s / %s", i.m.Category, i.m.Name) }

func (i scriptItem) Description() string {
	if i.m.Description == "" {
		return "(undocumented)"
	}
	return i.m.Description
}

func (i scriptItem) FilterValue() string {
	return strings.TrimSpace(i.m.Name + " " + i.m.Category + " " + i.m.Description)
}

// newItems builds list items sorted by (Category, Name) so that scripts
// belonging to the same category are grouped together in the list.
func newItems(metas []meta.ScriptMeta) []list.Item {
	sorted := make([]meta.ScriptMeta, len(metas))
	copy(sorted, metas)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Category != sorted[j].Category {
			return sorted[i].Category < sorted[j].Category
		}
		return sorted[i].Name < sorted[j].Name
	})

	items := make([]list.Item, 0, len(sorted))
	for _, m := range sorted {
		items = append(items, scriptItem{m: m})
	}
	return items
}
