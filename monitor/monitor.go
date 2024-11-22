package monitor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/fatih/color"

	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

type TokenInfo struct {
	Symbol    string
	Address   string
	CreatedAt string
	Score     int64
}

type TokenMeta struct {
	Name            string `json:"name"`
	Symbol          string `json:"symbol"`
	URI             string `json:"uri"`
	Mutable         bool   `json:"mutable"`
	UpdateAuthority string `json:"updateAuthority"`
}

type KnownAccounts map[string]struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Holder struct {
	Address        string  `json:"address"`
	Amount         int64   `json:"amount"`
	Decimals       int     `json:"decimals"`
	Pct            float64 `json:"pct"`
	UiAmount       float64 `json:"uiAmount"`
	UiAmountString string  `json:"uiAmountString"`
	Owner          string  `json:"owner"`
	Insider        bool    `json:"insider"`
}

type Report struct {
	TokenMeta            TokenMeta     `json:"tokenMeta"`
	Risks                []Risk        `json:"risks"`
	TotalMarketLiquidity float64       `json:"totalMarketLiquidity"`
	TotalLPProviders     int           `json:"totalLPProviders"`
	Rugged               bool          `json:"rugged"`
	KnownAccounts        KnownAccounts `json:"knownAccounts"`
	Verification         string        `json:"verification"`
	Score                int           `json:"score"`
	FreezeAuthority      string        `json:"freezeAuthority"`
	MintAuthority        string        `json:"mintAuthority"`
	TopHolders           []Holder      `json:"topHolders"`
}

type Risk struct {
	Name  string
	Score int64
	Level string
}

func Run() {
	slog.Info(color.New(color.BgHiBlue).SprintFunc()("Connecting to WebSocket..."))
	client, err := ConnectToWebSocket()
	if err != nil {
		slog.Error(color.New(color.BgBlack, color.FgRed).SprintFunc()(fmt.Sprintf("Failed to connect to WebSocket: %v", err)))
	}

	defer client.Close()

	pubkey := "7YttLkHDoNj9wyDur5pM1ejNaAvT9X4eqaYcHQqtj2G5" // ray_fee_pubkey
	slog.Info(color.New(color.BgHiCyan).SprintFunc()("Subscribing to logs..."))
	if err := SubscribeToLogs(client, pubkey); err != nil {
		slog.Error(color.New(color.BgBlack, color.FgRed).SprintFunc()(fmt.Sprintf("Failed to subscribe to logs: %v", err)))
	}
}

func displayTokenTable(tokens []TokenInfo) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"SYMBOL", "ADDRESS", "CREATED AT", "SCORE", "URL"})

	for _, token := range tokens {
		address := fmt.Sprintf("%s...%s", token.Address[:4], token.Address[len(token.Address)-4:])
		url := fmt.Sprintf("https://api.rugcheck.xyz/v1/tokens/%s/report", token.Address)
		table.Append([]string{token.Symbol, address, token.CreatedAt, fmt.Sprintf("%d", token.Score), url})
	}

	table.Render()
}

func checkMintAddress(mint string) (string, []Risk, error) {
	url := fmt.Sprintf("https://api.rugcheck.xyz/v1/tokens/%s/report", mint)
	var symbol string
	var risks []Risk
	var tokens []TokenInfo

	for attempts := 0; attempts < 3; attempts++ {
		resp, err := http.Get(url)
		if err != nil {
			slog.Error(color.New(color.BgBlack, color.FgRed).SprintFunc()(fmt.Sprintf("Error fetching data: %v", err)))
			return "", nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var report Report
			if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
				return "", nil, err
			}

			risks = report.Risks

			symbol = report.TokenMeta.Symbol

			createdAt := time.Now().Format("2006-01-02 15:04:05")
			for _, risk := range risks {
				token := TokenInfo{
					Symbol:    symbol,
					Address:   mint[:5] + "...",
					CreatedAt: createdAt,
					Score:     risk.Score,
				}
				tokens = append(tokens, token)
			}
			displayTokenTable(tokens)
			break
		}
		time.Sleep(5 * time.Second)
	}

	return symbol, risks, nil
}

// ConnectToWebSocket establishes a WebSocket connection to the Solana MainNet Beta.
func ConnectToWebSocket() (*ws.Client, error) {
	client, err := ws.Connect(context.Background(), "wss://mainnet.helius-rpc.com/?api-key=7bbbdbba-4a0f-4812-8112-757fbafbe571") // rpc.MainNetBeta_WS: "wss://api.mainnet-beta.solana.com" || wss://mainnet.helius-rpc.com/?api-key=7bbbdbba-4a0f-4812-8112-757fbafbe571
	if err != nil {
		return nil, err
	}
	return client, nil
}

func SubscribeToLogs(client *ws.Client, pubkey string) error {
	program := solana.MustPublicKeyFromBase58(pubkey)

	sub, err := client.LogsSubscribeMentions(
		program,
		rpc.CommitmentConfirmed,
	)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	slog.Info(color.New(color.BgHiGreen).SprintFunc()("Start monitoring..."))
	for {
		msg, err := sub.Recv(context.Background())
		if err != nil {
			return err
		}
		processLogMessage(msg)
	}
}

func processLogMessage(msg *ws.LogResult) {
	if msg.Value.Err != nil {
		slog.Error(color.New(color.BgBlack, color.FgRed).SprintFunc()(fmt.Sprintf("Transaction failed: %v", msg.Value.Err)))
		return
	}

	signature := msg.Value.Signature
	slog.Info(color.New(color.BgHiMagenta).SprintFunc()(fmt.Sprintf("Transaction Signature: %s", signature)))

	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=7bbbdbba-4a0f-4812-8112-757fbafbe571") // rpc.MainNetBeta_RPC: https://api.mainnet-beta.solana.com ||
	getTransactionDetails(rpcClient, signature)
}

func getTransactionDetails(rpcClient *rpc.Client, signature solana.Signature) {
	cero := uint64(0) // :/

	slog.Info(color.New(color.BgHiBlue).SprintFunc()("GetTransaction EncodingBase58..."))
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
		slog.Error(color.New(color.BgBlack, color.FgRed).SprintFunc()(fmt.Sprintf("Error GetTransaction EncodingBase58: %v", err)))
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

				// Check mint address for additional information
				symbol, risks, err := checkMintAddress(balance.Mint.String())
				if err != nil {
					slog.Error(color.New(color.BgBlack, color.FgRed).SprintFunc()(fmt.Sprintf("Error checking mint address: %v", err)))
				} else {
					slog.Info(color.New(color.BgHiGreen).SprintFunc()(fmt.Sprintf("Token Symbol: %s", symbol)))
					for _, risk := range risks {
						slog.Info(color.New(color.BgHiRed).SprintFunc()(fmt.Sprintf("Risk: %s, Score: %d, Level: %s", risk.Name, risk.Score, risk.Level)))
					}
				}
			}
		}
	}

}
