package ui

import (
	"fmt"
	"gosol/monitor"
	"gosol/types"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
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
	activeView    int
	table         table.Model
	statusBar     StatusListModel
	selectedToken *types.Report
	apiClient     *monitor.APIClient
	statusUpdates <-chan monitor.StatusMessage
	tokenUpdates  <-chan []types.TokenInfo
	stateManager  *monitor.StateManager
	stdoutView    StdoutViewModel

func NewModel(app *monitor.App) Model {
	columns := []table.Column{
		{Title: "", Width: 2},
		{Title: "CREATED AT", Width: 10},
		{Title: "SYMBOL", Width: 10},
		{Title: "SCORE", Width: 10},
		{Title: "ADDRESS", Width: 10},
		// {Title: "URL", Width: 100},
	}

	rows := []table.Row{}
	for _, token := range []types.TokenInfo{} {
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

	// messages := app.StateManager.GetStatusHistory()
	messages := []monitor.StatusMessage{}

	return Model{
		table:         t,
		statusBar:     NewStatusListModel(messages),
		statusUpdates: app.StatusUpdates,
		tokenUpdates:  app.TokenUpdates,
		stateManager:  app.StateManager,
		apiClient:     app.ApiClient,
		activeView:    1,
		stdoutView:    NewStdoutViewModel(),
	}
}

func (m *Model) updateTokenTable(tokens []types.TokenInfo) {
	m.statusBar.list.NewStatusMessage("Updating token table")
	rows := []table.Row{}
	for _, token := range tokens {
		if token.Address == "" && token.Symbol == "" && token.CreatedAt == "" && token.Score == 0 {
			continue
		}
		address := token.Address[:7] + "..."
		scoreColor := "🟢" // Green
		if token.Score > 2000 {
			scoreColor = "🟡" // Yellow
		}
		if token.Score > 3000 {
			scoreColor = "🟠" // Orange
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
		}
		rows = append(rows, row)
	}

	m.table.SetRows(rows)
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.stdoutView.Init())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Manejar cambio de tamaño
	case tea.KeyMsg:
		// Manejar entradas de teclado
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
				m.statusBar.list.CursorDown()
			}
		case "down":
			if m.activeView == 0 {
				if m.table.Cursor() < len(m.table.Rows())-1 {
					m.table.MoveDown(1)
				}
			} else if m.activeView == 1 {
				m.statusBar.list.CursorUp()
			}
		case "enter":
			// Obtener el token seleccionado de la tabla
			selectedRow := m.table.SelectedRow()
			if selectedRow != nil {
				// Asume que el primer campo es el símbolo del token
				symbol := selectedRow[2]

				ms := m.stateManager.GetMintState()
				for _, row := range ms {
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
	case monitor.StatusMessage:
		// Actualizar statusHistory en el UI
		// m.statusBar.list.SetItems(append(m.statusBar.list.Items(), listItem{message: msg}))
		// m.statusBar.list.NewStatusMessage(msg.Message)
		m.statusBar.list.InsertItem(0, listItem{message: msg})
	case TokenUpdateMsg:
		m.statusBar.list.NewStatusMessage("Received token update for: " + msg[0].Symbol)
		m.updateTokenTable(msg)
		// case StatusBarUpdateMsg:
		// 	messages := m.stateManager.GetStatusHistory()
		// 	// Limitar los mensajes a los últimos 10
		// 	if len(messages) > 10 {
		// 		messages = messages[len(messages)-10:]
		// 	}
		// 	items := make([]list.Item, len(messages))
		// 	for i, msg := range messages {
		// 		items[i] = listItem{message: msg}
		// 	}

		// 	// for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		// 	// 	items[i], items[j] = items[j], items[i]
		// 	// }

		// 	// Actualizar el modelo de la lista con los nuevos elementos
		// 	m.statusBar.list.SetItems(items)
	}

	cmd, _ := m.stdoutView.Update(msg)
	cmds = append(cmds, cmd)
	cmd, _ := m.statusBar.Update(msg)
	cmds = append(cmds, cmd)

	// Asegúrate de que el spinner se actualice en cada ciclo
	// spinnerCmd := m.statusBar.spinner.Tick
	// cmds = append(cmds, spinnerCmd)

	// Escuchar en los canales y enviar mensajes recibidos al modelo
	cmds = append(cmds,
		listenOnStatusUpdates(m.statusUpdates),
		listenOnTokenUpdates(m.tokenUpdates),
	)

	return m, tea.Batch(cmds...)
}

// esto envia un StatusMessage al Update verificar que lo reciba correctamente y ejecutar el Update
func listenOnStatusUpdates(ch <-chan monitor.StatusMessage) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

// esto envia un slice de tokens al Update verificar que lo reciba correctamente y ejecutar el Update
func listenOnTokenUpdates(ch <-chan []types.TokenInfo) tea.Cmd {
	return func() tea.Msg {
		tokens, ok := <-ch
		if !ok {
			return nil
		}
		return TokenUpdateMsg(tokens)
	}
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
	return fmt.Sprintf("\n%s\n%s\n%s", statusBarView, tableView, stdoutView)
}

func (m Model) tokenDetailView() string {
	if m.selectedToken == nil {
		return ""
	}
	if m.stateManager == nil {
		return "Error: StateManager not initialized."
	}

	// request update
	go m.apiClient.RequestReportOnDemand(m.selectedToken.Mint)

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
