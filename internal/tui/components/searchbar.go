package components

import "fmt"

// SearchBar renders the "/ search" prompt with the live query and the
// current match count, shown while the user is in search mode.
func SearchBar(query string, matchCount int) string {
	prompt := SearchPromptStyle.Render(" / search")
	q := SearchQueryStyle.Render(query)
	count := SearchCountStyle.Render(fmt.Sprintf("%d matches", matchCount))
	return prompt + "   " + q + "   " + count
}
