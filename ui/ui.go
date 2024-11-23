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
	table     table.Model
	statusBar string
}

func NewModel(tokens []types.TokenInfo) Model {
	columns := []table.Column{
		{Title: "", Width: 2},
		{Title: "CREATED AT", Width: 10},
		{Title: "SYMBOL", Width: 10},
		{Title: "SCORE", Width: 10},
		{Title: "ADDRESS", Width: 10},
		{Title: "URL", Width: 100},
	}

	rows := []table.Row{}
	for _, token := range tokens {
		if token.Address == "" && token.Symbol == "" && token.CreatedAt == "" && token.Score == 0 {
			continue
		}
		address := token.Address[:7] + "..."
		url := fmt.Sprintf("https://rugcheck.xyz/tokens/%s", token.Address)
		scoreColor := "🟢" // Green
		if token.Score > 2000 {
			scoreColor = "🟡" // Yellow
		}
		if token.Score > 3000 {
			scoreColor = "🟠" // Yellow
		}
		if token.Score > 4000 {
			scoreColor = "🔴" // Red
		}
		row := table.Row{
			scoreColor,
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
		table.WithHeight(7),
		table.WithFocused(true),
	)

	return Model{table: t}
}

func (m Model) Init() tea.Cmd {
	return nil
}

type TokenUpdateMsg []types.TokenInfo
type StatusBarUpdateMsg StatusMessage

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
		m.statusBar = formatStatusBar(msg)
	}
	return m, nil
}

func formatStatusBar(msg StatusMessage) string {
    var color lipgloss.Color
    switch msg.Level {
    case INFO:
        color = lipgloss.Color("2") // Green
    case WARN:
        color = lipgloss.Color("3") // Yellow
    case ERR:
        color = lipgloss.Color("1") // Red
    default:
        color = lipgloss.Color("241") // Gray
    }
    return lipgloss.NewStyle().Foreground(color).Render(msg.Message)
}
	tableView := m.table.View()
	statusBarView := formatStatusBar(StatusMessage{Level: NONE, Message: m.statusBar})
	return fmt.Sprintf("\n%s\n\n%s", statusBarView, tableView)
}
