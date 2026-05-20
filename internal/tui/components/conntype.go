package components

import "github.com/hech/mash/internal/config"

// TypeLabel returns the full human-readable label for the detail panel
// ("AWS EC2", "GCP Compute", ...) coloured by connection type.
func TypeLabel(t config.ConnType) string {
	switch t {
	case config.TypeEC2:
		return EC2Style.Render("AWS EC2")
	case config.TypeGCP:
		return GCPStyle.Render("GCP Compute")
	case config.TypeAzure:
		return AzureStyle.Render("Azure VM")
	case config.TypeTailscale:
		return TailscaleStyle.Render("Tailscale")
	case config.TypeMosh:
		return MoshStyle.Render("Mosh")
	default:
		return SSHStyle.Render("SSH")
	}
}

// StyleConnType renders the short type code ("SSH", "EC2", ...) used
// in the Type column of the connections table.
func StyleConnType(ct config.ConnType) string {
	s := string(ct)
	switch ct {
	case config.TypeSSH:
		return SSHStyle.Render(s)
	case config.TypeMosh:
		return MoshStyle.Render(s)
	case config.TypeEC2:
		return EC2Style.Render(s)
	case config.TypeGCP:
		return GCPStyle.Render(s)
	case config.TypeAzure:
		return AzureStyle.Render(s)
	case config.TypeTailscale:
		return TailscaleStyle.Render(s)
	default:
		return s
	}
}
