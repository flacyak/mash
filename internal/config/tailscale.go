package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
)

// TailscaleStatus mirrors the output of "tailscale status --json".
type TailscaleStatus struct {
	Peer map[string]TailscalePeer `json:"Peer"`
	Self TailscalePeer            `json:"Self"`
}

// TailscalePeer represents a device in the Tailscale mesh.
type TailscalePeer struct {
	HostName  string   `json:"HostName"`
	DNSName   string   `json:"DNSName"`
	TailscaleIPs []string `json:"TailscaleIPs"`
	OS        string   `json:"OS"`
	Online    bool     `json:"Online"`
	ExitNode  bool     `json:"ExitNode"`
}

// DiscoverTailscaleConnections parses "tailscale status --json" output and
// returns a Connection for each peer. If statusPath is set it reads from that
// file; otherwise it shells out to the tailscale CLI.
func DiscoverTailscaleConnections(statusPath string) ([]Connection, error) {
	var raw []byte
	var err error

	if statusPath != "" {
		raw, err = os.ReadFile(statusPath)
	} else {
		raw, err = exec.Command("tailscale", "status", "--json").Output()
	}
	if err != nil {
		return nil, fmt.Errorf("tailscale status: %w", err)
	}

	var status TailscaleStatus
	if err := json.Unmarshal(raw, &status); err != nil {
		return nil, fmt.Errorf("parse tailscale status: %w", err)
	}

	var conns []Connection

	for _, peer := range status.Peer {
		if peer.HostName == "" {
			continue
		}
		c := Connection{
			Name:   peer.HostName,
			Type:   TypeTailscale,
			Uptime: "-",
			User:   "-",
		}
		if len(peer.TailscaleIPs) > 0 {
			c.Host = peer.TailscaleIPs[0]
		}
		if peer.DNSName != "" {
			c.Host = peer.DNSName
		}
		// Tailscale SSH runs on port 22 by default, but we note port 443
		// to distinguish from local SSH entries.
		c.Port = "443"

		if peer.ExitNode {
			c.Name = c.Name + " [exit]"
		}
		conns = append(conns, c)
	}

	sort.Slice(conns, func(i, j int) bool {
		return conns[i].Name < conns[j].Name
	})

	return conns, nil
}
