package components

import (
	"fmt"
	"strings"
)

// FooterEntry is a single keybinding hint shown in the footer.
type FooterEntry struct {
	Key, Text string
}

// FooterItem is a small constructor so callers don't need to know
// FooterEntry's field names.
func FooterItem(key, text string) FooterEntry {
	return FooterEntry{Key: key, Text: text}
}

// Footer renders the bottom hint line: connection count, then each
// keybinding hint separated by a dim middle dot.
func Footer(count int, items ...FooterEntry) string {
	parts := []string{" " + FooterKeyStyle.Render(fmt.Sprintf("%d", count)) + "  " + FooterStyle.Render("connections")}
	for _, it := range items {
		parts = append(parts, FooterKeyStyle.Render(it.Key)+"  "+FooterStyle.Render(it.Text))
	}
	sep := FooterSepStyle.Render("  · ")
	return strings.Join(parts, sep)
}
