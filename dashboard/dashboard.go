package dashboard

import (
	"fmt"
	"sort"
	"strings"

	"gosol/monitor"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	mintState map[string][]monitor.Report
}

type tickMsg struct{}

func NewModel(mintState map[string][]monitor.Report) model {
	return model{mintState: mintState}
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return func() tea.Msg {
		return tickMsg{}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tickMsg:
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
		return true
	})

	b.WriteString("Mint Addresses:\n")
	for _, mint := range mints {
		b.WriteString(fmt.Sprintf("- %s\n", mint))
	}

	b.WriteString("\nReports:\n")
	for _, mint := range mints {
		reports := m.mintState[mint]
		for _, report := range reports {
			b.WriteString(fmt.Sprintf("Mint: %s\n", mint))
			b.WriteString(fmt.Sprintf("Symbol: %s\n", report.TokenMeta.Symbol))
			b.WriteString(fmt.Sprintf("Score: %d\n", report.Score))
			b.WriteString(fmt.Sprintf("Risks: %d\n", len(report.Risks)))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func RunDashboard(mintState map[string][]monitor.Report) {
	p := tea.NewProgram(NewModel(mintState))
	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
