package monitor

import (
	"context"
	"fmt"
	"gosol/types"
	"log/slog"

	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

func updateStatus(status string, statusUpdates chan<- string) {
	slog.Log(context.TODO(), slog.LevelInfo, fmt.Sprintf("%s", color.New(color.BgHiBlue).SprintFunc()(status)), time.Now().Format("15:04"))
}

var mintState = make(map[string][]types.Report)

func (m *Monitor) getMintState() map[string][]types.Report {
	return mintState
}

type Monitor struct {
	tokenUpdates  chan<- []types.TokenInfo
	statusUpdates chan<- string
}

func NewMonitor(tokenUpdates chan<- []types.TokenInfo, statusUpdates chan<- string) *Monitor {
	return &Monitor{tokenUpdates: tokenUpdates, statusUpdates: statusUpdates}
}

func (m *Monitor) Run() {
	updateStatus("Connecting to WebSocket...", m.statusUpdates)
	client, err := m.connectToWebSocket()
	if err != nil {
		updateStatus(fmt.Sprintf("Failed to connect to WebSocket: %v", err), m.statusUpdates)
	}

	defer client.Close()

	pubkey := "7YttLkHDoNj9wyDur5pM1ejNaAvT9X4eqaYcHQqtj2G5" // ray_fee_pubkey
	updateStatus("Subscribing to logs...", m.statusUpdates)
	if err := m.subscribeToLogs(client, pubkey); err != nil {
		updateStatus(fmt.Sprintf("Failed to subscribe to logs: %v", err), m.statusUpdates)
	}
}

type model struct {
	table table.Model
}

func newModel(tokens []types.TokenInfo) model {
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

	return model{table: t}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	return m.table.View()
}

func displayTokenTable(tokens []types.TokenInfo) {
	p := tea.NewProgram(newModel(tokens))
	if err := p.Start(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}

func (m *Monitor) checkMintAddress(mint string) (string, []types.Risk, error) {
	url := fmt.Sprintf("https://api.rugcheck.xyz/v1/tokens/%s/report", mint)
	var symbol string
	var risks []types.Risk
	var tokens []types.TokenInfo

	for attempts := 0; attempts < 3; attempts++ {
		resp, err := http.Get(url)
		if err != nil {
			// updateStatus(fmt.Sprintf("Error fetching data: %v", err))
			return "", nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var report types.Report
			if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
				return "", nil, err
			}

			risks = report.Risks

			symbol = report.TokenMeta.Symbol

			token := types.TokenInfo{
				Symbol:    symbol,
				Address:   mint,
				CreatedAt: report.DetectedAt.Format("15:04"),
				Score:     int64(report.Score),
			}
			tokens = append(tokens, token)
			// Update the in-memory state with the report
			mintState[mint] = []types.Report{report}

			// se envia el listado de tokens a la UI
			m.tokenUpdates <- tokens
			break
		}
		time.Sleep(5 * time.Second)
	}

	return symbol, risks, nil
}

func (m *Monitor) connectToWebSocket() (*ws.Client, error) {
	client, err := ws.Connect(context.Background(), "wss://mainnet.helius-rpc.com/?api-key=7bbbdbba-4a0f-4812-8112-757fbafbe571") // rpc.MainNetBeta_WS: "wss://api.mainnet-beta.solana.com" || wss://mainnet.helius-rpc.com/?api-key=7bbbdbba-4a0f-4812-8112-757fbafbe571
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (m *Monitor) subscribeToLogs(client *ws.Client, pubkey string) error {
	program := solana.MustPublicKeyFromBase58(pubkey)

	sub, err := client.LogsSubscribeMentions(
		program,
		rpc.CommitmentConfirmed,
	)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	// updateStatus("Start monitoring...")
	for {
		msg, err := sub.Recv(context.Background())
		if err != nil {
			return err
		}
		m.processLogMessage(msg)
	}
}

func (m *Monitor) processLogMessage(msg *ws.LogResult) {
	if msg.Value.Err != nil {
		// updateStatus(fmt.Sprintf("Transaction failed: %v", msg.Value.Err))
		return
	}

	signature := msg.Value.Signature
	// updateStatus(fmt.Sprintf("Transaction Signature: %s", signature))

	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=7bbbdbba-4a0f-4812-8112-757fbafbe571") // rpc.MainNetBeta_RPC: https://api.mainnet-beta.solana.com ||
	m.getTransactionDetails(rpcClient, signature)
}

func (m *Monitor) getTransactionDetails(rpcClient *rpc.Client, signature solana.Signature) {
	cero := uint64(0) // :/

	// updateStatus("GetTransaction EncodingBase58...")
	tx58, err := rpcClient.GetTransaction(
		context.TODO(),
		signature,
		&rpc.GetTransactionOpts{
			Encoding:                       solana.EncodingBase58,
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: &cero,
		},
	)
	if err != nil {
		// updateStatus(fmt.Sprintf("Error GetTransaction EncodingBase58: %v", err))
	}

	if tx58 != nil {
		// spew.Dump(tx58)
		// spew.Dump(tx58.Transaction.GetBinary())
		// if txJson.Meta != nil {
		for _, balance := range tx58.Meta.PostTokenBalances {
			if balance.Owner.String() == "5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1" && balance.Mint.String() != "So11111111111111111111111111111111111111112" {
				// slog.Info(color.New(color.BgHiYellow).SprintFunc()("========== New Token Found =========="))
				// slog.Info(color.New(color.BgHiYellow).SprintFunc()(fmt.Sprintf("Mint Address: %s", balance.Mint)))
				// slog.Info(color.New(color.BgHiYellow).SprintFunc()("====================================="))

				// Add mint address to mintState if it doesn't exist
				if _, exists := mintState[balance.Mint.String()]; !exists {
					mintState[balance.Mint.String()] = []types.Report{}
				}

				// Check mint address for additional API information
				go m.checkMintAddress(balance.Mint.String())
			}
		}
	}
}
