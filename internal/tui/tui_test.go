package tui

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/hech/mash/internal/config"
)

var update = flag.Bool("update", false, "update golden files")

func mockConnections() []config.Connection {
	return []config.Connection{
		{
			Name:   "prod-web-01",
			Port:   "22",
			Type:   config.TypeSSH,
			Host:   "10.0.1.50",
			User:   "deploy",
			Uptime: "PT120H30M",
		},
		{
			Name:   "staging-db",
			Port:   "5432",
			Type:   config.TypeSSH,
			Host:   "db.staging.example.com",
			User:   "admin",
			Uptime: "< 1h",
		},
		{
			Name:   "bastion",
			Port:   "2222",
			Type:   config.TypeSSH,
			Host:   "bastion.corp.example.com",
			User:   "ops",
			Uptime: "-",
		},
		{
			Name:   "mosh-server (pid 54321)",
			Port:   "60001",
			Type:   config.TypeMosh,
			Host:   "localhost",
			Pid:    "54321",
			Uptime: "PT4H15M",
		},
	}
}

func smModel() Model {
	m := NewModel()
	conns := mockConnections()
	m.conns = conns
	LoadRows(&m)
	return m
}

// step sends a tea.Msg to the model and returns the updated Model plus any command.
func step(m Model, msg tea.Msg) (Model, tea.Cmd) {
	nm, cmd := m.Update(msg)
	return nm.(Model), cmd
}

func goldenPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("testdata", "goldens", name+".golden")
}

func stripAnsi(s string) string {
	var b strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inEscape {
			if ch == 'm' {
				inEscape = false
			}
			continue
		}
		if ch == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			continue
		}
		b.WriteByte(ch)
	}
	return b.String()
}

func diffLines(a, b string) string {
	al := strings.Split(a, "\n")
	bl := strings.Split(b, "\n")
	maxLen := len(al)
	if len(bl) > maxLen {
		maxLen = len(bl)
	}
	var out strings.Builder
	for i := 0; i < maxLen; i++ {
		var la, lb string
		if i < len(al) {
			la = al[i]
		}
		if i < len(bl) {
			lb = bl[i]
		}
		if la != lb {
			out.WriteString(fmt.Sprintf("  line %d:\n    want: %q\n    got:  %q\n", i+1, la, lb))
		}
	}
	return out.String()
}

func assertGolden(t *testing.T, name string, got string) {
	t.Helper()
	path := goldenPath(t, name)
	got = stripAnsi(got)

	if *update {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden %s: %v", path, err)
		}
		t.Logf("updated golden: %s", name)
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v\nRun with -update to create golden files.", path, err)
	}

	wantStr := string(want)
	if got != wantStr {
		t.Fatalf("golden mismatch for %s:\n--- want\n+++ got\n%s",
			name, diffLines(wantStr, got))
	}
}

func TestSmokeNavigationAndScreens(t *testing.T) {
	m := smModel()
	m, _ = step(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	// Screen 1: Browser view, row 0 (prod-web-01) selected.
	assertGolden(t, "browser_initial", m.View())

	// Navigate down twice to row 2 (bastion).
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	assertGolden(t, "browser_row3", m.View())

	// Enter selection with 'l' on bastion.
	var cmd tea.Cmd
	m, cmd = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if cmd == nil {
		t.Fatal("expected ping command after entering selection")
	}
	m, _ = step(m, pingResultMsg{ms: "12.5ms"})
	assertGolden(t, "detail_row3", m.View())

	// Navigate down to row 3 (mosh-server) while selected.
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = step(m, pingResultMsg{err: "unreachable"})
	assertGolden(t, "detail_row4_mosh", m.View())

	// Navigate up twice: first back to bastion, then to staging-db.
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m, _ = step(m, pingResultMsg{ms: "14.2ms"})
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m, _ = step(m, pingResultMsg{ms: "3.1ms"})
	assertGolden(t, "detail_row2_ssh", m.View())

	// Leave selection with 'h'.
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	assertGolden(t, "browser_after_select", m.View())

	// Quit with 'q'.
	m, cmd = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestEmptyState(t *testing.T) {
	m := NewModel()
	m, _ = step(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	assertGolden(t, "empty_state", m.View())
}

func TestArrowKeysAliases(t *testing.T) {
	m := smModel()
	m, _ = step(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	// down arrow equals j
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyDown})
	// left arrow equals h (no-op when not selected)
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyLeft})
	if m.selected {
		t.Fatal("left arrow should not toggle selection when not in selection mode")
	}
	if m.table.Cursor() != 1 {
		t.Fatalf("expected cursor at row 1, got %d", m.table.Cursor())
	}

	// right arrow equals l
	var cmd tea.Cmd
	m, cmd = step(m, tea.KeyMsg{Type: tea.KeyRight})
	if cmd == nil {
		t.Fatal("expected ping command after right arrow selection")
	}
	if !m.selected {
		t.Fatal("right arrow should enter selection mode")
	}
	m, _ = step(m, pingResultMsg{ms: "8.0ms"})

	// up arrow equals k
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyUp})
	m, _ = step(m, pingResultMsg{ms: "5.0ms"})
	if m.table.Cursor() != 0 {
		t.Fatalf("expected cursor at row 0 after up, got %d", m.table.Cursor())
	}

	// left arrow equals h (exit selection)
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyLeft})
	if m.selected {
		t.Fatal("left arrow should exit selection mode")
	}
}

func TestQuitWithoutConnections(t *testing.T) {
	m := NewModel()
	m, _ = step(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	_, cmd := step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("q should quit even with no connections")
	}
}
