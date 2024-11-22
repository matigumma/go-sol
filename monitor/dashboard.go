package monitor

import (
	"fmt"
	"sort"
	"strings"

	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func UpdateStatus(status string) {
	// Aquí puedes implementar la lógica para actualizar el estado en el dashboard
}

type model struct {
	mintState map[string][]Report
	status    string
}

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	mintStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	scoreStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("229"))
	riskStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("160"))
	statusStyle = lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("229")).Bold(true)
)

func (m *model) setStatus(status string) {
	m.status = status
}

type tickMsg struct{}

func NewModel(mintState map[string][]Report) model {
	return model{mintState: mintState}
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(1 * time.Second) // Simulate some work
		return tickMsg{}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tickMsg:
		m.setStatus("Updating dashboard...")
		time.Sleep(1 * time.Second) // Simulate some work
		m.setStatus("Dashboard updated.")
		return m, tickCmd()
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	mints := make([]string, 0, len(m.mintState))
	for mint := range m.mintState {
		mints = append(mints, mint)
	}
	sort.Slice(mints, func(i, j int) bool {
		return mints[i] < mints[j]
	})

	b.WriteString(titleStyle.Render("Mint Addresses:\n"))
	for _, mint := range mints {
		b.WriteString(mintStyle.Render(fmt.Sprintf("- %s\n", mint)))
	}

	b.WriteString(titleStyle.Render("\nReports:\n"))
	for _, mint := range mints {
		reports := m.mintState[mint]
		for _, report := range reports {
			b.WriteString(mintStyle.Render(fmt.Sprintf("Mint: %s\n", mint)))
			b.WriteString(fmt.Sprintf("Symbol: %s\n", report.TokenMeta.Symbol))
			b.WriteString(scoreStyle.Render(fmt.Sprintf("Score: %d\n", report.Score)))
			b.WriteString(riskStyle.Render(fmt.Sprintf("Risks: %d\n", len(report.Risks))))
			b.WriteString("\n")
		}
	}

	return b.String() + "\n" + statusStyle.Render(m.status)
}

func RunDashboard(mintState map[string][]Report) {
	p := tea.NewProgram(NewModel(mintState), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
