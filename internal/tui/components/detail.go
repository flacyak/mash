package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/hech/mash/internal/config"
)

// DetailPanel renders the right-hand panel that appears once a row is
// selected: type label, ping status, connection metadata, stored
// credential indicator, and any recorded "Common Issues" for the host.
func DetailPanel(c config.Connection, pingMs string, pinging bool, issues config.HostIssues, hasVaulted bool) string {
	if c.User == "" {
		c.User = "-"
	}

	typeLine := " " + TypeLabel(c.Type)
	rule := SectionRuleStyle.Render(strings.Repeat("─", SectionRuleLen))

	var pingLine string
	switch {
	case pinging:
		pingLine = PingWait.Render(" ⠿ pinging…")
	case pingMs == "":
		pingLine = PingWait.Render(" · awaiting")
	case pingMs == "unreachable" || pingMs == "no response":
		pingLine = PingFail.Render(" ✗ " + pingMs)
	default:
		pingLine = PingOk.Render(" ✓ " + pingMs)
	}

	lines := []string{
		"",
		" " + typeLine,
		"",
		SectionHeaderStyle.Render(" Status"),
		" " + rule,
		pingLine,
		"",
		SectionHeaderStyle.Render(" Connection"),
		" " + rule,
		" " + KeyStyle.Render(fmt.Sprintf("%-4s", "host")) + " " + ValueStyle.Render(c.Host),
		" " + KeyStyle.Render(fmt.Sprintf("%-4s", "user")) + " " + ValueStyle.Render(c.User),
		" " + KeyStyle.Render(fmt.Sprintf("%-4s", "port")) + " " + ValueStyle.Render(c.Port),
	}
	if c.Pid != "" {
		lines = append(lines, " "+KeyStyle.Render(fmt.Sprintf("%-4s", "pid"))+" "+ValueStyle.Render(c.Pid))
	}
	if c.Uptime != "" && c.Uptime != "-" {
		lines = append(lines, " "+KeyStyle.Render(fmt.Sprintf("%-4s", "up"))+" "+ValueStyle.Render(c.Uptime))
	}
	if hasVaulted {
		lines = append(lines, " "+KeyStyle.Render(fmt.Sprintf("%-4s", "auth"))+" "+PingOk.Render("stored"))
	}

	if issueLines := renderIssueLines(issues, rule); len(issueLines) > 0 {
		lines = append(lines, "")
		lines = append(lines, issueLines...)
	}
	lines = append(lines, "")

	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return DetailPanelStyle.Render(body)
}

// issueCategoryOrder fixes the display order so the panel layout
// doesn't flip between renders (Go map iteration is randomised).
var issueCategoryOrder = []config.IssueCategory{
	config.IssueTimeout,
	config.IssueRefused,
	config.IssueNetwork,
}

var issueCategoryLabels = map[config.IssueCategory]string{
	config.IssueTimeout: "Connection Timeout",
	config.IssueRefused: "Connection Refused",
	config.IssueNetwork: "Network Blocking",
}

func renderIssueLines(issues config.HostIssues, rule string) []string {
	if len(issues) == 0 {
		return nil
	}

	out := []string{
		SectionHeaderStyle.Render(" Common Issues"),
		" " + rule,
	}

	seen := make(map[config.IssueCategory]bool, len(issues))
	first := true
	emit := func(cat config.IssueCategory, reasons []string) {
		if len(reasons) == 0 {
			return
		}
		if !first {
			out = append(out, "")
		}
		first = false
		label, ok := issueCategoryLabels[cat]
		if !ok {
			label = strings.Title(string(cat))
		}
		out = append(out, " "+IssueCategoryStyle.Render(label))
		for _, r := range reasons {
			out = append(out, "  "+IssueReasonStyle.Render("! "+r))
		}
	}

	for _, cat := range issueCategoryOrder {
		if reasons, ok := issues[cat]; ok {
			emit(cat, reasons)
			seen[cat] = true
		}
	}
	for cat, reasons := range issues {
		if !seen[cat] {
			emit(cat, reasons)
		}
	}
	return out
}
