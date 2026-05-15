# AGENTS.md

## Overview

**Mash** is a terminal TUI (text user interface) for managing SSH, Mosh, cloud VM, and Tailscale connections. It discovers connections from local SSH configs, running Mosh servers, OpenTofu cloud state, and the Tailscale mesh, then presents them in a browsable table with ping-based reachability and a detail panel.

- **Language:** Go 1.26
- **Module:** `github.com/hech/mash`
- **TUI framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Bubbles table](https://github.com/charmbracelet/bubbles) + [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **Build system:** Nix flake + `go build`

## Directory layout

```
mash/
  main.go                     # Entry point: tea.NewProgram(Model, tea.WithAltScreen())
  go.mod / go.sum
  flake.nix                   # Nix: packages.mash, checks.{mash-e2e, mash-golden-tests, mash-cloud-tests}, devShell
  AGENTS.md
  internal/
    config/                   # Connection discovery (no TUI dependency)
      types.go                #   ConnType enum, Connection struct
      ssh.go                  #   ParseSSHConfig, LoadAllSSHConnections, uptime detection
      mosh.go                 #   DiscoverMoshConnections (pgrep + UFW + ss)
      cloud.go                #   DiscoverCloudConnections (OpenTofu state JSON parser)
      tailscale.go            #   DiscoverTailscaleConnections (tailscale status --json)
    tui/
      tui.go                  #   Bubble Tea Model, Update, View, styling, search
      tui_test.go             #   Unit tests + golden files (smoke, empty, arrows, quit, search)
      tui_integration_test.go #   Integration test: fake ~/.ssh/config, real LoadAllSSHConnections
      tui_cloud_test.go       #   Cloud test: fake SSH + OpenTofu state + Tailscale status
      testdata/
        ssh_config            #   Fake SSH config with 5 hosts
        tofu_state.json       #   Pre-generated OpenTofu state (2 EC2, 2 GCP, 2 Azure)
        tailscale_status.json #   Pre-generated Tailscale status (3 peers)
        infra/main.tf         #   OpenTofu project to regenerate tofu_state.json
        goldens/              #   31 golden files for visual regression
```

## Architecture

### config package — no TUI dependency

All connection discovery lives in `internal/config/`. Every discovery function returns `[]Connection` and never touches the Bubble Tea model directly.

`Connection` is the universal type:

```go
type ConnType string  // SSH | Mosh | EC2 | GCP | Azure | Tailscale
type Connection struct {
    Name, Port, Host, User, Pid, Uptime string
    Type ConnType
}
```

**Discovery sources (in order):**

1. `LoadAllSSHConnections()` — parses `~/.ssh/config` and `/etc/ssh/ssh_config`, then calls `detectActiveSSHUptime()` to annotate active SSH sessions with PID and uptime.
2. `DiscoverMoshConnections()` — probes `pgrep mosh-server`, UFW rules (`/etc/ufw/*.rules`), and `ss -ulpn`; deduplicates by `port:host`.
3. `DiscoverCloudConnections(statePath)` — parses OpenTofu `tofu show -json` output. When `statePath` is empty, shells out to the `tofu` CLI. Recognizes `aws_instance`, `google_compute_instance`, `azurerm_linux_virtual_machine`, `azurerm_windows_virtual_machine`.
4. `DiscoverTailscaleConnections(statusPath)` — parses `tailscale status --json` output. Marks exit nodes with `[exit]` suffix in the name.

**Critical design rule:** All discovery functions accept a file-path parameter for testing. When the parameter is non-empty they read from the file; when empty they shell out. This lets integration tests run without real CLI tools installed.

### tui package — Bubble Tea application

**Model state machine:**

```
┌──────────┐   /    ┌──────────┐
│ Browser  │───────>│  Search  │
│ selected │<───────│  mode    │
│ = false  │ esc/r  │          │
└────┬─────┘        └──────────┘
     │ l/right
     v
┌──────────┐
│ Detail   │
│ Panel    │  (ping runs as tea.Cmd in background)
│ selected │
│ = true   │ h/left back to Browser
└──────────┘
```

**Key bindings:**

| Key | Context | Action |
|-----|---------|--------|
| `q`, `ctrl+c` | Anywhere | Quit |
| `j`, `down` | Browser | Move cursor down (fires ping in detail mode) |
| `k`, `up` | Browser | Move cursor up (fires ping in detail mode) |
| `l`, `right` | Browser | Enter detail panel, start ping |
| `h`, `left` | Detail panel | Back to browser |
| `/` | Browser | Open search mode |
| (typing) | Search | Append character, fuzzy-filter connections |
| `backspace` | Search | Remove last character, re-filter |
| `enter` | Search | Commit filtered results, exit search |
| `esc` or `right` | Search | Cancel and restore full connection list |

**Search** uses a simple sequential fuzzy match (`fuzzyMatch`): each query character must appear in order within the target name or host (case-insensitive). The full connection list is saved to `Model.allConns` when search opens and restored on cancel.

**Ping** runs via `tea.Cmd` (not a goroutine). It shells out to `ping -c 1 -W 1 <host>` and returns a `pingResultMsg` with either `ms` or an `err` string. The `pinging` flag controls the "Ping: ..." spinner in the detail panel.

**Styling:** All Lipgloss styles are package-level `var` blocks. Each connection type has its own color: SSH=blue, Mosh=pink, EC2=orange, GCP=cyan, Azure=dark-cyan, Tailscale=purple. The detail panel shows the provider label for cloud/Tailscale connections.

## Testing

### Golden file pattern

Every UI test uses **ANSI-stripped golden files**. The `assertGolden` function:
- Strips ANSI escape sequences from the rendered view
- Compares line-by-line against a `.golden` file in `testdata/goldens/`
- Supports `go test -update` to regenerate all golden files

```go
// Rebuild all golden files:
go test ./internal/tui/ -update -v

// Verify golden files (without -update):
go test ./internal/tui/ -v

// Run a single test:
go test ./internal/tui/ -run TestSearchFunctionality -v
```

### Test categories

| Test file | Tests | What it covers |
|-----------|-------|----------------|
| `tui_test.go` | `TestSmokeNavigationAndScreens`, `TestEmptyState`, `TestArrowKeysAliases`, `TestQuitWithoutConnections`, `TestSearchFunctionality`, `TestSearchOnEmpty`, `TestSearchBackspace` | Mocked connections, full navigation, search UX |
| `tui_integration_test.go` | `TestRealConfigNavigationAndScreens` | Sets `HOME` to a temp dir with fake `~/.ssh/config`, calls real `LoadAllSSHConnections()` |
| `tui_cloud_test.go` | `TestCloudNavigationAndScreens` | Fake SSH config + pre-generated OpenTofu state + Tailscale status = 14 total connections |

### The `step` helper

```go
func step(m Model, msg tea.Msg) (Model, tea.Cmd) {
    nm, cmd := m.Update(msg)
    return nm.(Model), cmd  // type-asserts to value type
}
```

**Important:** `Model.Update` returns `(tea.Model, tea.Cmd)`. The `step` helper type-asserts back to `Model` (value type, not pointer). All handler methods that use pointer receivers (like `handleSearchKey`) must return `*m` (dereference) so `step` can re-wrap.

### Test data files

| File | Purpose |
|------|---------|
| `testdata/ssh_config` | 5 fake hosts (db, AWS bastion, Redis, Grafana, nginx) |
| `testdata/tofu_state.json` | 6 cloud VMs (2 EC2, 2 GCP, 2 Azure) |
| `testdata/tailscale_status.json` | 3 Tailscale peers (home-server, office-nas, vpn-gateway) |
| `testdata/infra/main.tf` | OpenTofu project to regenerate the JSON state file |

## Nix

- **Build:** `nix build .#mash`
- **Dev shell:** `nix develop` (provides Go, gopls, goimports, OpenTofu)
- **Checks:**
  - `mash-e2e` — binary smoke test
  - `mash-golden-tests` — golden regression with fake SSH config
  - `mash-cloud-tests` — golden regression with cloud/Tailscale data

## Conventions

### Adding a new connection source

1. Add a new `Type*` constant in `config/types.go`.
2. Add a discovery function in `config/` (accept `filePath string` for testability).
3. Add a color style in `tui/tui.go`'s var block.
4. Add a case in `styleConnType()` and `renderDetailPanel()`.
5. Wire the call into `LoadConnections()`.
6. Create test data + an integration test + golden files.
7. Run `go test ./internal/tui/ -update` to regenerate goldens.

### Adding a new TUI mode

1. Add state fields to `Model`.
2. Handle the entering/exiting keys in `Update`.
3. Add a branch in `View()` for the new mode's layout.
4. Add golden snapshots covering transitions.
5. Regenerate golden files.

### File editing

The codebase uses **tabs** for indentation, not spaces. The `Edit` tool sometimes fails on tab-indented files — when that happens, rewrite the entire file with `Create` instead.

### Commit style

Short imperative subject lines (e.g., "syncing cloud connections", "base init").
