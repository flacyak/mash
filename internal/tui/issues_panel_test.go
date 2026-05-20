package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/hech/mash/internal/config"
)

func TestIssuesPanelRenders(t *testing.T) {
	m := smModel()
	m.issues = config.IssuesState{
		Hosts: map[string]config.HostIssues{
			"prod-web-01": {
				config.IssueTimeout: {"cloud firewall", "wrong IP address"},
				config.IssueRefused: {"sshd not running"},
				config.IssueNetwork: {"public wifi blocks outbound"},
			},
		},
	}
	m, _ = step(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m, _ = step(m, keyRune('l'))
	m, _ = step(m, pingResultMsg{ms: "12.5ms"})

	out := stripAnsi(viewStr(m))
	for _, want := range []string{
		"Common Issues",
		"Connection Timeout",
		"! cloud firewall",
		"! wrong IP address",
		"Connection Refused",
		"! sshd not running",
		"Network Blocking",
		"! public wifi blocks outbound",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q\nrendered:\n%s", want, out)
		}
	}
}

func TestIssuesPanelHiddenWhenEmpty(t *testing.T) {
	m := smModel()
	m, _ = step(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m, _ = step(m, keyRune('l'))
	m, _ = step(m, pingResultMsg{ms: "12.5ms"})

	out := stripAnsi(viewStr(m))
	if strings.Contains(out, "Common Issues") {
		t.Errorf("Common Issues section should not appear when host has no entries\nrendered:\n%s", out)
	}
}
