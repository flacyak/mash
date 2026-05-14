package tui

import (
	"fmt"
	"os/exec"
	"regexp"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/hech/mash/internal/config"
)

var (
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(0, 1)

	sshStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Bold(true)
	moshStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Bold(true)

	detailPanelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(1, 2).
				MarginLeft(1)

	sectionHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true).
				Underline(true).
				MarginBottom(1)

	keyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("229"))
	pingOk     = lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)
	pingFail   = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	pingWait   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

type Model struct {
	table    table.Model
	conns    []config.Connection
	width    int
	height   int
	selected bool
	pingMs   string
	pinging  bool
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
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
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

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

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

func (m Model) View() string {
	title := titleStyle.Render("Mash - Mosh/SSH Connection Manager")

	tableContent := baseStyle.Render(m.table.View())

	if !m.selected {
		count := len(m.table.Rows())
		hint := "j/k or arrows: navigate | l/right: select | q: quit"
		if count == 0 {
			hint = "No connections found | q: quit"
		}
		footer := footerStyle.Render(fmt.Sprintf(" %d connections | %s", count, hint))
		return lipgloss.JoinVertical(lipgloss.Left, title, tableContent, footer)
	}

	idx := m.table.Cursor()
	if idx >= len(m.conns) {
		footer := footerStyle.Render(fmt.Sprintf(" %d connections | h/left: back | j/k: change | q: quit", len(m.table.Rows())))
		return lipgloss.JoinVertical(lipgloss.Left, title, tableContent, footer)
	}

	c := m.conns[idx]
	detailContent := renderDetailPanel(c, m.pingMs, m.pinging)
	footer := footerStyle.Render(fmt.Sprintf(" %d connections | h/left: back | j/k: change selection | q: quit", len(m.table.Rows())))

	return lipgloss.JoinVertical(lipgloss.Left, title,
		lipgloss.JoinHorizontal(lipgloss.Top, tableContent, detailContent),
		footer,
	)
}

func renderDetailPanel(c config.Connection, pingMs string, pinging bool) string {
	if c.User == "" {
		c.User = "-"
	}

	statusSection := sectionHeaderStyle.Render("Status")
	var pingLine string
	if pinging {
		pingLine = pingWait.Render("Ping: ...")
	} else if pingMs == "" {
		pingLine = pingWait.Render("Ping: -")
	} else if pingMs == "unreachable" || pingMs == "no response" {
		pingLine = pingFail.Render("Ping: " + pingMs)
	} else {
		pingLine = pingOk.Render("Ping: " + pingMs)
	}

	connSection := sectionHeaderStyle.Render("Connection")
	hostLine := keyStyle.Render("host: ") + valueStyle.Render(c.Host)
	userLine := keyStyle.Render("user: ") + valueStyle.Render(c.User)
	portLine := keyStyle.Render("port: ") + valueStyle.Render(c.Port)

	body := lipgloss.JoinVertical(lipgloss.Left,
		statusSection,
		pingLine,
		"",
		connSection,
		hostLine,
		userLine,
		portLine,
	)

	return detailPanelStyle.Width(36).Render(body)
}

func LoadRows(m *Model) {
	rows := make([]table.Row, 0, len(m.conns))
	for _, c := range m.conns {
		typeStr := string(c.Type)
		if c.Type == config.TypeSSH {
			typeStr = sshStyle.Render(typeStr)
		} else {
			typeStr = moshStyle.Render(typeStr)
		}
		rows = append(rows, table.Row{c.Name, c.Port, typeStr, c.Host, c.Uptime})
	}
	m.table.SetRows(rows)
}

func LoadConnections(m *Model) {
	conns := config.LoadAllSSHConnections()
	moshConns := config.DiscoverMoshConnections()
	conns = append(conns, moshConns...)

	m.conns = conns
	LoadRows(m)
}
