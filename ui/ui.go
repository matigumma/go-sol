package ui

import (
	"fmt"
	"gosol/monitor"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render

type Model struct {
	table table.Model
}

func NewModel(tokens []monitor.TokenInfo) Model {
	columns := []table.Column{
		{Title: "SYMBOL", Width: 20},
		{Title: "ADDRESS", Width: 10},
		{Title: "CREATED AT", Width: 20},
		{Title: "SCORE", Width: 10},
		{Title: "URL", Width: 40},
	}

	rows := []table.Row{}
	seenAddresses := make(map[string]bool)
	for _, token := range tokens {
		if !seenAddresses[token.Address] {
			address := token.Address[:7]
			url := fmt.Sprintf("https://rugcheck.xyz/tokens/%s", token.Address)
			scoreColor := lipgloss.Color("2") // Green
			if token.Score > 2000 {
				scoreColor = lipgloss.Color("3") // Yellow
			}
			if token.Score > 4000 {
				scoreColor = lipgloss.Color("1") // Red
			}
			row := table.Row{token.Symbol, address, token.CreatedAt, lipgloss.NewStyle().Foreground(scoreColor).Render(fmt.Sprintf("%d", token.Score)), url}
			rows = append(rows, row)
			seenAddresses[token.Address] = true
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	return Model{table: t}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	return m.table.View()
}
