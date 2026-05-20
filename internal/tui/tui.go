package tui

import (
	"os/exec"
	"regexp"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/hech/mash/internal/config"
	"github.com/hech/mash/internal/tui/components"

	"github.com/hech/mash/internal/vault"
)

type Model struct {
	table       table.Model
	conns       []config.Connection
	allConns    []config.Connection
	issues      config.IssuesState
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
		Foreground(lipgloss.Color(components.MainColor)).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color(components.MainColor)).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color(components.HexToANSIString("#1292b4"))).
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
	title := components.TitleCard()
	tableContent := components.BaseStyle.Render(m.table.View())

	if m.searching {
		searchBar := components.SearchBar(m.searchQuery, len(m.conns))
		footer := components.Footer(len(m.conns),
			components.FooterItem("type", "filter"),
			components.FooterItem("enter", "keep"),
			components.FooterItem("esc", "cancel"),
		)
		return lipgloss.JoinVertical(lipgloss.Left, title, searchBar, tableContent, footer)
	}

	if !m.selected {
		if len(m.conns) == 0 {
			hint := components.EmptyHintStyle.Render("  No connections discovered yet.")
			footer := components.Footer(0, components.FooterItem("q", "quit"))
			return lipgloss.JoinVertical(lipgloss.Left, title, tableContent, "", hint, "", footer)
		}
		footer := components.Footer(len(m.conns),
			components.FooterItem("j/k", "nav"),
			components.FooterItem("l", "select"),
			components.FooterItem("/", "search"),
			components.FooterItem("q", "quit"),
		)
		return lipgloss.JoinVertical(lipgloss.Left, title, tableContent, footer)
	}

	idx := m.table.Cursor()
	footer := components.Footer(len(m.conns),
		components.FooterItem("h", "back"),
		components.FooterItem("j/k", "switch"),
		components.FooterItem("q", "quit"),
	)
	if idx >= len(m.conns) {
		return lipgloss.JoinVertical(lipgloss.Left, title, tableContent, footer)
	}

	c := m.conns[idx]
	detailContent := components.DetailPanel(c, m.pingMs, m.pinging, m.issues.For(c.Name), vault.Has(c.Name))

	return lipgloss.JoinVertical(lipgloss.Left, title,
		lipgloss.JoinHorizontal(lipgloss.Top, tableContent, detailContent),
		footer,
	)
}

func LoadRows(m *Model) {
	rows := make([]table.Row, 0, len(m.conns))
	for _, c := range m.conns {
		typeStr := components.StyleConnType(c.Type)
		rows = append(rows, table.Row{c.Name, c.Port, typeStr, c.Host, c.Uptime})
	}
	m.table.SetRows(rows)
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
	m.issues, _ = config.LoadIssuesState(config.DefaultIssuesPath())
	LoadRows(m)
}

func LoadWithTestData(m *Model, tofuPath, tailscalePath string) error {
	conns := config.LoadAllSSHConnections()
	moshConns := config.DiscoverMoshConnections()
	conns = append(conns, moshConns...)

	cloudConns, err := config.DiscoverCloudConnections(tofuPath)
	if err != nil {
		return err
	}
	conns = append(conns, cloudConns...)

	tailConns, err := config.DiscoverTailscaleConnections(tailscalePath)
	if err != nil {
		return err
	}
	conns = append(conns, tailConns...)

	m.conns = conns
	m.issues, _ = config.LoadIssuesState(config.DefaultIssuesPath())
	LoadRows(m)
	return nil
}
