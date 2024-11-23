package ui

import (
	"fmt"
	"gosol/types"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render

type Model struct {
	table table.Model
	statusBar string
}

func NewModel(tokens []types.TokenInfo) Model {
	columns := []table.Column{
		{Title: "CREATED AT", Width: 25},
		{Title: "SYMBOL", Width: 10},
		{Title: "SCORE", Width: 10},
		{Title: "ADDRESS", Width: 15},
		{Title: "URL", Width: 50},
	}

	rows := []table.Row{}
	for _, token := range tokens {
		if token.Address == "" && token.Symbol == "" && token.CreatedAt == "" && token.Score == 0 {
			continue
		}
		address := token.Address[:7] + "..."
		url := fmt.Sprintf("https://rugcheck.xyz/tokens/%s", token.Address)
		// scoreColor := color.BgGreen // Green
		// if token.Score > 2000 {
		// 	scoreColor = color.BgYellow // Yellow
		// }
		// if token.Score > 4000 {
		// 	scoreColor = color.BgRed // Red
		// }
		row := table.Row{
			token.CreatedAt,
			token.Symbol,
			fmt.Sprintf("%d", token.Score),
			address,
			url,
		}
		rows = append(rows, row)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(5),
	)

	return Model{table: t}
}

func (m Model) Init() tea.Cmd {
	return nil
}

type TokenUpdateMsg []types.TokenInfo
type StatusBarUpdateMsg string

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case TokenUpdateMsg:
		// Actualiza la tabla con los nuevos tokens
		m.table = NewModel(msg).table
	case StatusBarUpdateMsg:
		// Actualiza la ui de statusbar
		// Actualiza la barra de estado con el mensaje recibido
		m.statusBar = string(msg)
	}
	return m, nil
}

func (m Model) View() string {
	tableView := m.table.View()
	statusBarView := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(m.statusBar)
	return fmt.Sprintf("%s\n\n%s", tableView, statusBarView)
}
