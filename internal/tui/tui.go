package tui

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/hech/mash/internal/config"
)

const (
	titleCardWidth   = 48
	detailPanelWidth = 36
	sectionRuleLen   = 24
)

var (
	baseStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))

	titleCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	titleBrandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	titleSubStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	footerKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	footerSepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	sshStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Bold(true)
	moshStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Bold(true)
	ec2Style       = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	gcpStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	azureStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
	tailscaleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Bold(true)

	detailPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(0, 1).
				MarginLeft(1).
				Width(detailPanelWidth)

	sectionHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)

	sectionRuleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	keyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("229"))
	pingOk     = lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)
	pingFail   = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	pingWait   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	emptyHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)

	searchPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)

	searchQueryStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("229")).
				Bold(true)

	searchCountStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245"))
)

type Model struct {
	table       table.Model
	conns       []config.Connection
	allConns    []config.Connection
	width       int
	height      int
	selected    bool
	pingMs      string
	pinging     bool
	searching   bool
	searchQuery string
}

type pingResultMsg struct {
	ms  string
	err string
}

func pingCmd(host string) tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("ping", "-c", "1", "-W", "1", host).Output()
		if err != nil {
			return pingResultMsg{err: "unreachable"}
		}
		re := regexp.MustCompile(`time=(\d+\.?\d*)\s*ms`)
		matches := re.FindStringSubmatch(string(out))
		if len(matches) > 1 {
			return pingResultMsg{ms: matches[1] + "ms"}
		}
		return pingResultMsg{err: "no response"}
	}
}

func NewModel() Model {
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Port", Width: 8},
		{Title: "Type", Width: 8},
		{Title: "Host", Width: 30},
		{Title: "Uptime", Width: 16},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(20),
		table.WithWidth(120),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		Foreground(lipgloss.Color("245")).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("240")).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return Model{table: t}
}

func (m *Model) resizeTable() {
	tableWidth := m.width - 4
	if m.selected {
		tableWidth = m.width - 40
	}
	cols := m.table.Columns()
	cols[0].Width = max(10, tableWidth-72)
	cols[3].Width = max(10, tableWidth-72)
	m.table.SetColumns(cols)

	innerWidth := 0
	for _, c := range cols {
		innerWidth += c.Width + 2
	}
	m.table.SetWidth(innerWidth)
	m.table.SetHeight(m.height - 6)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeTable()
		return m, nil

	case pingResultMsg:
		m.pinging = false
		if msg.err != "" {
			m.pingMs = msg.err
		} else {
			m.pingMs = msg.ms
		}
		return m, nil

	case tea.KeyPressMsg:
		if m.searching {
			return m.handleSearchKey(msg)
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "/":
			if len(m.conns) > 0 && !m.selected {
				m.searching = true
				m.searchQuery = ""
				m.allConns = append([]config.Connection{}, m.conns...)
				return m, nil
			}
			return m, nil

		case "l", "right":
			if !m.selected && len(m.conns) > 0 {
				m.selected = true
				m.pingMs = ""
				m.pinging = true
				m.resizeTable()
				c := m.conns[m.table.Cursor()]
				return m, pingCmd(c.Host)
			}
			return m, nil

		case "h", "left":
			if m.selected {
				m.selected = false
				m.pingMs = ""
				m.pinging = false
				m.resizeTable()
			}
			return m, nil

		case "j", "down":
			m.table.MoveDown(1)
			if m.selected {
				idx := m.table.Cursor()
				if idx < len(m.conns) {
					c := m.conns[idx]
					m.pingMs = ""
					m.pinging = true
					return m, pingCmd(c.Host)
				}
			}
			return m, nil

		case "k", "up":
			m.table.MoveUp(1)
			if m.selected {
				idx := m.table.Cursor()
				if idx < len(m.conns) {
					c := m.conns[idx]
					m.pingMs = ""
					m.pinging = true
					return m, pingCmd(c.Host)
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *Model) handleSearchKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.searching = false
		return *m, nil

	case "esc":
		m.searching = false
		m.conns = m.allConns
		LoadRows(m)
		m.table.SetCursor(0)
		return *m, nil

	case "l", "right":
		m.searching = false
		m.conns = m.allConns
		LoadRows(m)
		m.table.SetCursor(0)
		return *m, nil

	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.filterConns()
		}
		return *m, nil

	case "space", "tab", "left", "up", "down", "ctrl+c":
		return *m, nil

	default:
		if msg.Text != "" {
			m.searchQuery += msg.Text
			m.filterConns()
		}
		return *m, nil
	}
}

func (m *Model) filterConns() {
	if m.searchQuery == "" {
		m.conns = m.allConns
	} else {
		filtered := make([]config.Connection, 0)
		for _, c := range m.allConns {
			if fuzzyMatch(m.searchQuery, c.Name) || fuzzyMatch(m.searchQuery, c.Host) {
				filtered = append(filtered, c)
			}
		}
		m.conns = filtered
	}
	LoadRows(m)
	if len(m.conns) > 0 {
		m.table.SetCursor(0)
	}
}

func fuzzyMatch(query, target string) bool {
	q := strings.ToLower(query)
	t := strings.ToLower(target)
	qi := 0
	for i := 0; i < len(t) && qi < len(q); i++ {
		if t[i] == q[qi] {
			qi++
		}
	}
	return qi == len(q)
}

func (m Model) View() tea.View {
	v := tea.NewView(m.renderScreen())
	v.AltScreen = true
	return v
}

func (m Model) renderScreen() string {
	title := renderTitleCard()

	if m.searching {
		searchBar := renderSearchBar(m.searchQuery, len(m.conns))
		tableContent := baseStyle.Render(m.table.View())
		footer := renderFooter(len(m.conns),
			footerItem("type", "filter"),
			footerItem("enter", "keep"),
			footerItem("esc", "cancel"),
		)
		return lipgloss.JoinVertical(lipgloss.Left, title, searchBar, tableContent, footer)
	}

	tableContent := baseStyle.Render(m.table.View())

	if !m.selected {
		if len(m.conns) == 0 {
			hint := emptyHintStyle.Render("  No connections discovered yet.")
			footer := renderFooter(0, footerItem("q", "quit"))
			return lipgloss.JoinVertical(lipgloss.Left, title, tableContent, "", hint, "", footer)
		}
		footer := renderFooter(len(m.conns),
			footerItem("j/k", "nav"),
			footerItem("l", "select"),
			footerItem("/", "search"),
			footerItem("q", "quit"),
		)
		return lipgloss.JoinVertical(lipgloss.Left, title, tableContent, footer)
	}

	idx := m.table.Cursor()
	if idx >= len(m.conns) {
		footer := renderFooter(len(m.conns),
			footerItem("h", "back"),
			footerItem("j/k", "switch"),
			footerItem("q", "quit"),
		)
		return lipgloss.JoinVertical(lipgloss.Left, title, tableContent, footer)
	}

	c := m.conns[idx]
	detailContent := renderDetailPanel(c, m.pingMs, m.pinging)
	footer := renderFooter(len(m.conns),
		footerItem("h", "back"),
		footerItem("j/k", "switch"),
		footerItem("q", "quit"),
	)

	return lipgloss.JoinVertical(lipgloss.Left, title,
		lipgloss.JoinHorizontal(lipgloss.Top, tableContent, detailContent),
		footer,
	)
}

func renderTitleCard() string {
	brand := titleBrandStyle.Render("MASH")
	sub := titleSubStyle.Render("Mosh / SSH / Cloud Connection Manager")
	content := " " + brand + "  " + sub
	return titleCardStyle.Render(content)
}

type footerEntry struct {
	key, text string
}

func footerItem(key, text string) footerEntry {
	return footerEntry{key: key, text: text}
}

func renderFooter(count int, items ...footerEntry) string {
	parts := []string{" " + footerKeyStyle.Render(fmt.Sprintf("%d", count)) + "  " + footerStyle.Render("connections")}
	for _, it := range items {
		parts = append(parts, footerKeyStyle.Render(it.key)+"  "+footerStyle.Render(it.text))
	}
	sep := footerSepStyle.Render("  · ")
	return strings.Join(parts, sep)
}

func renderSearchBar(query string, matchCount int) string {
	prompt := searchPromptStyle.Render(" / search")
	q := searchQueryStyle.Render(query)
	count := searchCountStyle.Render(fmt.Sprintf("%d matches", matchCount))
	return prompt + "   " + q + "   " + count
}

func renderDetailPanel(c config.Connection, pingMs string, pinging bool) string {
	if c.User == "" {
		c.User = "-"
	}

	typeLine := " " + renderTypeLabel(c.Type)

	rule := sectionRuleStyle.Render(strings.Repeat("─", sectionRuleLen))

	var pingLine string
	switch {
	case pinging:
		pingLine = pingWait.Render(" ⠿ pinging…")
	case pingMs == "":
		pingLine = pingWait.Render(" · awaiting")
	case pingMs == "unreachable" || pingMs == "no response":
		pingLine = pingFail.Render(" ✗ " + pingMs)
	default:
		pingLine = pingOk.Render(" ✓ " + pingMs)
	}

	lines := []string{
		"",
		" " + typeLine,
		"",
		sectionHeaderStyle.Render(" Status"),
		" " + rule,
		pingLine,
		"",
		sectionHeaderStyle.Render(" Connection"),
		" " + rule,
		" " + keyStyle.Render(fmt.Sprintf("%-4s", "host")) + " " + valueStyle.Render(c.Host),
		" " + keyStyle.Render(fmt.Sprintf("%-4s", "user")) + " " + valueStyle.Render(c.User),
		" " + keyStyle.Render(fmt.Sprintf("%-4s", "port")) + " " + valueStyle.Render(c.Port),
	}
	if c.Pid != "" {
		lines = append(lines, " "+keyStyle.Render(fmt.Sprintf("%-4s", "pid"))+" "+valueStyle.Render(c.Pid))
	}
	if c.Uptime != "" && c.Uptime != "-" {
		lines = append(lines, " "+keyStyle.Render(fmt.Sprintf("%-4s", "up"))+" "+valueStyle.Render(c.Uptime))
	}
	lines = append(lines, "")

	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return detailPanelStyle.Render(body)
}

func renderTypeLabel(t config.ConnType) string {
	switch t {
	case config.TypeEC2:
		return ec2Style.Render("AWS EC2")
	case config.TypeGCP:
		return gcpStyle.Render("GCP Compute")
	case config.TypeAzure:
		return azureStyle.Render("Azure VM")
	case config.TypeTailscale:
		return tailscaleStyle.Render("Tailscale")
	case config.TypeMosh:
		return moshStyle.Render("Mosh")
	default:
		return sshStyle.Render("SSH")
	}
}

func LoadRows(m *Model) {
	rows := make([]table.Row, 0, len(m.conns))
	for _, c := range m.conns {
		typeStr := styleConnType(c.Type)
		rows = append(rows, table.Row{c.Name, c.Port, typeStr, c.Host, c.Uptime})
	}
	m.table.SetRows(rows)
}

func styleConnType(ct config.ConnType) string {
	s := string(ct)
	switch ct {
	case config.TypeSSH:
		return sshStyle.Render(s)
	case config.TypeMosh:
		return moshStyle.Render(s)
	case config.TypeEC2:
		return ec2Style.Render(s)
	case config.TypeGCP:
		return gcpStyle.Render(s)
	case config.TypeAzure:
		return azureStyle.Render(s)
	case config.TypeTailscale:
		return tailscaleStyle.Render(s)
	default:
		return s
	}
}

func LoadConnections(m *Model) {
	conns := config.LoadAllSSHConnections()
	moshConns := config.DiscoverMoshConnections()
	conns = append(conns, moshConns...)

	cloudConns, _ := config.DiscoverCloudConnections("")
	conns = append(conns, cloudConns...)

	tailConns, _ := config.DiscoverTailscaleConnections("")
	conns = append(conns, tailConns...)

	m.conns = conns
	LoadRows(m)
}
