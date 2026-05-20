package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// IssueCategory groups a recorded issue under one of the well-known
// diagnostic buckets ("timeout", "refused", "network"). Unknown keys
// are preserved verbatim and rendered with their raw name capitalized,
// so users can extend the state file without code changes.
type IssueCategory string

const (
	IssueTimeout IssueCategory = "timeout"
	IssueRefused IssueCategory = "refused"
	IssueNetwork IssueCategory = "network"
)

// HostIssues maps each category to the list of reason strings the user
// has recorded for that host. JSON-serialised on disk under the host
// name in IssuesState.Hosts.
type HostIssues map[IssueCategory][]string

// IssuesState is the root of the user state file (XDG_DATA_HOME/mash/state.json).
// Hosts is keyed by Connection.Name so SSH config entries, cloud tag names
// and Tailscale hostnames all resolve through the same lookup.
type IssuesState struct {
	Hosts map[string]HostIssues `json:"hosts"`
}

// For returns the recorded issues for the given connection name, or
// nil if the host has no entry. Safe to call on a zero-value state.
func (s IssuesState) For(name string) HostIssues {
	if s.Hosts == nil {
		return nil
	}
	return s.Hosts[name]
}

// DefaultIssuesPath returns the canonical XDG state file path:
// $XDG_DATA_HOME/mash/state.json, falling back to $HOME/.local/share/mash/state.json.
// Returns an empty string only if neither variable resolves to a directory
// (e.g. HOME unset in an unusual environment).
func DefaultIssuesPath() string {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, "mash", "state.json")
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".local", "share", "mash", "state.json")
	}
	return ""
}

// LoadIssuesState reads and parses the state file at path. A missing
// file is not an error — it just yields an empty state, since most
// users will run mash before ever creating one.
func LoadIssuesState(path string) (IssuesState, error) {
	if path == "" {
		return IssuesState{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return IssuesState{}, nil
		}
		return IssuesState{}, err
	}
	var s IssuesState
	if err := json.Unmarshal(data, &s); err != nil {
		return IssuesState{}, err
	}
	return s, nil
}
