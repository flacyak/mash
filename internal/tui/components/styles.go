// Package components holds pure rendering helpers for the mash TUI:
// title card, footer, search bar, connection-type label, and the
// per-connection detail panel. None of these touch model state — they
// take primitive arguments and return rendered strings — so they can
// be exercised individually and recombined by the top-level View.
package components

import "charm.land/lipgloss/v2"

// Palette constants. Numbers are 256-colour codes; named here so the
// same value is used by the row column, the type label, and the detail
// panel without drift.
const (
	DetailPanelWidth = 36
	SectionRuleLen   = 24

	MainColor      = "240"
	GroundColor    = "86"
	SubColor       = "245"
	SSHColor       = "75"
	MoshColor      = "240"
	EC2Color       = "214"
	GCPColor       = "39"
	AzureColor     = "33"
	TailscaleColor = "141"
	IssueColor     = "203"
)

var (
	BaseStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(MainColor))

	TitleCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(MainColor)).
			Padding(0, 1)

	TitleBrandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(GroundColor)).
			Bold(true)

	TitleSubStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(SubColor))

	FooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(SubColor))

	FooterKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(GroundColor)).
			Bold(true)

	FooterSepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(MainColor))

	SSHStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color(SSHColor)).Bold(true)
	MoshStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(MoshColor)).Bold(true)
	EC2Style       = lipgloss.NewStyle().Foreground(lipgloss.Color(EC2Color)).Bold(true)
	GCPStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color(GCPColor)).Bold(true)
	AzureStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(AzureColor)).Bold(true)
	TailscaleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(TailscaleColor)).Bold(true)

	DetailPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(MainColor)).
				Padding(0, 1).
				MarginLeft(1).
				Width(DetailPanelWidth)

	SectionHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(GroundColor)).
				Bold(true)

	SectionRuleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(MainColor))

	KeyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(SubColor))
	ValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("229"))
	PingOk     = lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)
	PingFail   = lipgloss.NewStyle().Foreground(lipgloss.Color(MoshColor))
	PingWait   = lipgloss.NewStyle().Foreground(lipgloss.Color(SubColor))

	EmptyHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(SubColor)).
			Italic(true)

	SearchPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(GroundColor)).
				Bold(true)

	SearchQueryStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("229")).
				Bold(true)

	SearchCountStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(MainColor))

	IssueCategoryStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(SubColor))

	IssueReasonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(IssueColor)).
				Bold(true)
)
