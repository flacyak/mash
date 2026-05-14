package tui

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/hech/mash/internal/config"
)

// cloudModel creates a Model loaded with SSH config connections + cloud
// connections from the test data tofu state and tailscale status files.
func cloudModel(t *testing.T) Model {
	t.Helper()

	m := NewModel()

	conns := config.LoadAllSSHConnections()
	moshConns := config.DiscoverMoshConnections()
	conns = append(conns, moshConns...)

	cloudConns, err := config.DiscoverCloudConnections(
		filepath.Join("testdata", "tofu_state.json"),
	)
	if err != nil {
		t.Fatalf("DiscoverCloudConnections: %v", err)
	}
	conns = append(conns, cloudConns...)

	tailConns, err := config.DiscoverTailscaleConnections(
		filepath.Join("testdata", "tailscale_status.json"),
	)
	if err != nil {
		t.Fatalf("DiscoverTailscaleConnections: %v", err)
	}
	conns = append(conns, tailConns...)

	m.conns = conns
	LoadRows(&m)
	return m
}

func TestCloudNavigationAndScreens(t *testing.T) {
	cleanup := setupFakeHome(t)
	defer cleanup()

	m := cloudModel(t)
	// 5 SSH + 6 cloud (2 EC2 + 2 GCP + 2 Azure) + 3 tailscale = 14
	if len(m.conns) != 14 {
		t.Fatalf("expected 14 total connections, got %d", len(m.conns))
	}

	m, _ = step(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	// Screen 1: Full browser with all 14 connections, row 0 (prod-db-primary).
	assertGolden(t, "cloud_browser_initial", m.View())

	// Navigate to row 5 (first EC2: ec2-prod-web-us-east).
	for i := 0; i < 5; i++ {
		m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}
	assertGolden(t, "cloud_browser_first_ec2", m.View())

	// Enter selection on first EC2.
	var cmd tea.Cmd
	m, cmd = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if cmd == nil {
		t.Fatal("expected ping command")
	}
	m, _ = step(m, pingResultMsg{ms: "42.1ms"})
	assertGolden(t, "cloud_detail_ec2", m.View())

	// Navigate down 2 to first GCP (gcp-data-processor).
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = step(m, pingResultMsg{ms: "18.7ms"})
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = step(m, pingResultMsg{ms: "22.3ms"})
	assertGolden(t, "cloud_detail_gcp", m.View())

	// Navigate down 2 to first Azure (az-sql-server-01).
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = step(m, pingResultMsg{ms: "95.8ms"})
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = step(m, pingResultMsg{ms: "101.2ms"})
	assertGolden(t, "cloud_detail_azure", m.View())

	// Navigate down 2 to first Tailscale (home-server).
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = step(m, pingResultMsg{ms: "1.2ms"})
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = step(m, pingResultMsg{ms: "1.5ms"})
	assertGolden(t, "cloud_detail_tailscale", m.View())

	// Leave selection.
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	assertGolden(t, "cloud_browser_after_select", m.View())

	// Quit.
	m, cmd = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}
