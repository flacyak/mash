package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"github.com/hech/mash/internal/tui"
)

//go:embed testdata/ssh_config
var sshConfigData []byte

//go:embed testdata/tofu_state.json
var tofuStateData []byte

//go:embed testdata/tailscale_status.json
var tailscaleStatusData []byte

//go:embed testdata/state.json
var stateData []byte

// main launches the mash TUI populated from embedded fixtures (mirrored
// from internal/tui/testdata) so the VHS demo always shows a
// deterministic 14-connection list: 5 SSH + 6 cloud + 3 Tailscale,
// regardless of the working directory the binary is run from.
func main() {
	tmpHome, err := os.MkdirTemp("", "mash-demo-home-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "tempdir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpHome)

	sshDir := filepath.Join(tmpHome, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir ssh: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "config"), sshConfigData, 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "write ssh config: %v\n", err)
		os.Exit(1)
	}
	os.Setenv("HOME", tmpHome)

	tofuPath := filepath.Join(tmpHome, "tofu_state.json")
	if err := os.WriteFile(tofuPath, tofuStateData, 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "write tofu state: %v\n", err)
		os.Exit(1)
	}

	tailPath := filepath.Join(tmpHome, "tailscale_status.json")
	if err := os.WriteFile(tailPath, tailscaleStatusData, 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "write tailscale status: %v\n", err)
		os.Exit(1)
	}

	// Place the issues state file at the XDG-default location so the
	// detail panel surfaces "Common Issues" for the demo connection.
	xdgData := filepath.Join(tmpHome, ".local", "share")
	stateDir := filepath.Join(xdgData, "mash")
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir state: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "state.json"), stateData, 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "write state: %v\n", err)
		os.Exit(1)
	}
	os.Setenv("XDG_DATA_HOME", xdgData)

	m := tui.NewModel()
	if err := tui.LoadWithTestData(&m, tofuPath, tailPath); err != nil {
		fmt.Fprintf(os.Stderr, "load test data: %v\n", err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
