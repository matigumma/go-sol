package ui

import (
	"fmt"
	"gosol/monitor"
	"gosol/types"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	activeBorderStyle   = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("205"))
	inactiveBorderStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render

type TokenUpdateMsg []types.TokenInfo
type StatusBarUpdateMsg monitor.StatusMessage

type Model struct {
	activeView int
	table      table.Model
	// statusBar      string
	statusBar      StatusListModel
	selectedToken  *types.Report
	currentMonitor *monitor.Monitor
}

func NewModel(tokens []types.TokenInfo) Model {
	columns := []table.Column{
		{Title: "", Width: 2},
		{Title: "CREATED AT", Width: 10},
		{Title: "SYMBOL", Width: 10},
		{Title: "SCORE", Width: 10},
		{Title: "ADDRESS", Width: 10},
		// {Title: "URL", Width: 100},
	}

	rows := []table.Row{}
	for _, token := range tokens {
		if token.Address == "" && token.Symbol == "" && token.CreatedAt == "" && token.Score == 0 {
			continue
		}
		address := token.Address[:7] + "..."
		// url := fmt.Sprintf("https://rugcheck.xyz/tokens/%s", token.Address)
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
			// url,
		}
		rows = append(rows, row)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	return Model{table: t}
}

func InitProject(monitor *monitor.Monitor) (tea.Model, tea.Cmd) {
	// Inicializar el modelo de la tabla con tokens vacíos
	m := NewModel([]types.TokenInfo{})
	m.currentMonitor = monitor
	m.activeView = 0

	// Obtener el historial de mensajes de estado y crear el modelo de lista
	messages := monitor.GetStatusHistory()
	m.statusBar = NewStatusListModel(messages)

	// Comando inicial para Bubble Tea, si es necesario
	cmd := tea.Batch(
	// Aquí puedes agregar comandos iniciales si los necesitas
	)

	return m, cmd
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.table.SetWidth(msg.Width - h)
		m.statusBar.list.SetSize(msg.Width-h, msg.Height-v)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.activeView = (m.activeView + 1) % 2
		case "up":
			if m.activeView == 0 {
				if m.table.Cursor() > 0 {
					m.table.MoveUp(1)
				}
			} else if m.activeView == 1 {
				m.statusBar.list.CursorUp()
			}
		case "down":
			if m.activeView == 0 {
				if m.table.Cursor() < len(m.table.Rows())-1 {
					m.table.MoveDown(1)
				}
			} else if m.activeView == 1 {
				m.statusBar.list.CursorDown()
			}
		case "enter":
			// Obtener el token seleccionado de la tabla
			selectedRow := m.table.SelectedRow()
			if selectedRow != nil {
				// Asume que el primer campo es el símbolo del token
				symbol := selectedRow[2]

				mintState := m.currentMonitor.GetMintState()

				for _, row := range mintState {
					if row[0].TokenMeta.Symbol == symbol { // Accede al índice correcto
						m.selectedToken = &row[0]
						break
					}
				}
			}
		case "esc":
			// Volver a la vista de la tabla
			m.selectedToken = nil
		}
	case TokenUpdateMsg:
		m.table = NewModel(msg).table
	case StatusBarUpdateMsg:
		messages := m.currentMonitor.GetStatusHistory()
		// Convertir los mensajes en list.Items
		items := make([]list.Item, len(messages))
		for i, msg := range messages {
			items[i] = listItem{message: msg}
		}

		for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
			items[i], items[j] = items[j], items[i]
		}

		// Actualizar el modelo de la lista con los nuevos elementos
		m.statusBar.list.SetItems(items)
	}

	// Actualizar el spinner
	cmd, _ := m.statusBar.Update(msg)
	cmds = append(cmds, cmd)

	// Asegúrate de que el spinner se actualice en cada ciclo
	spinnerCmd := m.statusBar.spinner.Tick
	cmds = append(cmds, spinnerCmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.selectedToken != nil {
		return m.tokenDetailView()
	}
	var tableView, statusBarView string

	// Apply active or inactive border style based on activeView
	if m.activeView == 0 {
		tableView = activeBorderStyle.Render(m.table.View())
		statusBarView = inactiveBorderStyle.Render(m.statusBar.View())
	} else {
		tableView = inactiveBorderStyle.Render(m.table.View())
		statusBarView = activeBorderStyle.Render(m.statusBar.View())
	}
	return fmt.Sprintf("\n%s\n%s", statusBarView, tableView)
}

func (m Model) tokenDetailView() string {
	if m.selectedToken == nil {
		return ""
	}
	if m.currentMonitor == nil {
		return "Error: Monitor not initialized."
	}

	// request update
	go m.currentMonitor.CheckMintAddress(m.selectedToken.Mint)

	time.Sleep(1 * time.Second)

	markdownContent := formatReportAsMarkdown(*m.selectedToken)

	// Usar glamour para renderizar el Markdown
	renderedContent, err := glamour.Render(markdownContent, "dark")
	if err != nil {
		return fmt.Sprintf("Error rendering markdown: %v", err)
	}

	return renderedContent
}

func formatReportAsMarkdown(report types.Report) string {
	var risks []string
	for _, risk := range report.Risks {
		risks = append(risks, fmt.Sprintf("- **%s**: %s (Score: %d)", risk.Name, risk.Level, risk.Score))
	}

	return fmt.Sprintf(`                                                                                                                                                 
# Token Report: %s                                                                                                                                                       
																																										 
**Symbol**: %s                                                                                                                                                           
**Name**: %s                                                                                                                                                             
**Score**: %d                                                                                                                                                            
**Rugged**: %t                                                                                                                                                           
**Verification**: %s                                                                                                                                                     
																																										 
## Known Accounts                                                                                                                                                        
%s                                                                                                                                                                       
																																										 
## Risks                                                                                                                                                                 
%s                                                                                                                                                                       
																																										 
## Top Holders                                                                                                                                                           
%s                                                                                                                                                                       
`,
		report.TokenMeta.Name,
		report.TokenMeta.Symbol,
		report.TokenMeta.Name,
		report.Score,
		report.Rugged,
		report.Verification,
		formatKnownAccounts(report.KnownAccounts),
		strings.Join(risks, "\n"),
		formatTopHolders(report.TopHolders),
	)
}

func formatKnownAccounts(accounts types.KnownAccounts) string {
	var result []string
	for address, account := range accounts {
		result = append(result, fmt.Sprintf("- **%s**: %s (%s)", address, account.Name, account.Type))
	}
	return strings.Join(result, "\n")
}

func formatTopHolders(holders []types.Holder) string {
	var result []string
	for _, holder := range holders {
		result = append(result, fmt.Sprintf("- **%s**: %d (%.2f%%)", holder.Address, holder.Amount, holder.Pct))
	}
	return strings.Join(result, "\n")
}

// func formatStatusBar(msg StatusBarUpdateMsg) string {
// 	var color lipgloss.Color
// 	switch msg.Level {
// 	case monitor.INFO:
// 		color = lipgloss.Color("2") // Green
// 	case monitor.WARN:
// 		color = lipgloss.Color("3") // Yellow
// 	case monitor.ERR:
// 		color = lipgloss.Color("1") // Red
// 	default:
// 		color = lipgloss.Color("241") // Gray
// 	}
// 	return lipgloss.NewStyle().Foreground(color).Render(msg.Message)
// }

func parseScore(scoreStr string) int64 {
	score, err := strconv.ParseInt(scoreStr, 10, 64)
	if err != nil {
		return 0
	}
	return score
}
