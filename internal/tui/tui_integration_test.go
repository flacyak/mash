package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/hech/mash/internal/config"
)

// setupFakeHome creates a temporary HOME directory containing a .ssh/config
// copied from testdata, sets the HOME environment variable, and returns
// a cleanup function that restores HOME and removes the temp dir.
func setupFakeHome(t *testing.T) func() {
	t.Helper()

	oldHome := os.Getenv("HOME")

	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatalf("mkdir .ssh: %v", err)
	}

	src, err := os.ReadFile(filepath.Join("testdata", "ssh_config"))
	if err != nil {
		t.Fatalf("read testdata ssh_config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "config"), src, 0o600); err != nil {
		t.Fatalf("write ssh config: %v", err)
	}

	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("setenv HOME: %v", err)
	}

	return func() {
		os.Setenv("HOME", oldHome)
	}
}

// realConfigModel creates a Model populated via real config loading from the
// fake HOME directory (set up by setupFakeHome).
func realConfigModel(t *testing.T) Model {
	t.Helper()

	m := NewModel()
	conns := config.LoadAllSSHConnections()

	moshConns := config.DiscoverMoshConnections()
	conns = append(conns, moshConns...)

	m.conns = conns
	LoadRows(&m)
	return m
}

func TestRealConfigNavigationAndScreens(t *testing.T) {
	cleanup := setupFakeHome(t)
	defer cleanup()

	m := realConfigModel(t)
	if len(m.conns) != 5 {
		t.Fatalf("expected 5 connections from test SSH config, got %d", len(m.conns))
	}

	m, _ = step(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	// Screen 1: Browser view with 5 real connections, row 0 selected.
	assertGolden(t, "real_browser_initial", m.View())

	// Navigate down twice to row 2 (redis-cache-01).
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	assertGolden(t, "real_browser_row3", m.View())

	// Enter selection with 'l' on redis-cache-01.
	var cmd tea.Cmd
	m, cmd = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if cmd == nil {
		t.Fatal("expected ping command after entering selection")
	}
	m, _ = step(m, pingResultMsg{ms: "0.8ms"})
	assertGolden(t, "real_detail_row3", m.View())

	// Navigate down to row 3 (grafana-monitoring) while selected.
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = step(m, pingResultMsg{ms: "2.3ms"})
	assertGolden(t, "real_detail_row4", m.View())

	// Navigate up twice: back to redis, then to aws-bastion-us-east.
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m, _ = step(m, pingResultMsg{ms: "1.1ms"})
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m, _ = step(m, pingResultMsg{ms: "15.4ms"})
	assertGolden(t, "real_detail_row2", m.View())

	// Leave selection with 'h'.
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	assertGolden(t, "real_browser_after_select", m.View())

	// Quit with 'q'.
	m, cmd = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}
