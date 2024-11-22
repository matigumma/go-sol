package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"github.com/fatih/color"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

func Run() {
	slog.Info(color.New(color.BgGreen).SprintFunc()("Connecting to WebSocket..."))
	client, err := ConnectToWebSocket()
	if err != nil {
		slog.Error(color.New(color.BgRed).SprintFunc()(fmt.Sprintf("Failed to connect to WebSocket: %v", err)))
	}
	defer client.Close()

	pubkey := "7YttLkHDoNj9wyDur5pM1ejNaAvT9X4eqaYcHQqtj2G5" // ray_fee_pubkey
	slog.Info(color.New(color.BgGreen).SprintFunc()("Subscribing to logs..."))
	if err := SubscribeToLogs(client, pubkey); err != nil {
		slog.Error(color.New(color.BgRed).SprintFunc()(fmt.Sprintf("Failed to subscribe to logs: %v", err)))
	}
}

// ConnectToWebSocket establishes a WebSocket connection to the Solana MainNet Beta.
func ConnectToWebSocket() (*ws.Client, error) {
	client, err := ws.Connect(context.Background(), rpc.MainNetBeta_WS) // "wss://api.mainnet-beta.solana.com" || wss://mainnet.helius-rpc.com/?api-key=7bbbdbba-4a0f-4812-8112-757fbafbe571
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

	slog.Info(color.New(color.BgGreen).SprintFunc()("Start monitoring..."))
	for {
		msg, err := sub.Recv(context.Background())
		if err != nil {
			return err
		}
		processLogMessage(msg)
	}
}

func processLogMessage(msg *ws.LogResult) {
	signature := msg.Value.Signature
	slog.Info(color.New(color.BgBlue).SprintFunc()(fmt.Sprintf("Transaction Signature: %s", signature)))

	rpcClient := rpc.New(rpc.MainNetBeta_RPC)
	getTransactionDetails(rpcClient, signature.String())
}

func getTransactionDetails(rpcClient *rpc.Client, signature string) {
	cero := uint64(0) // :/

	tx, err := rpcClient.GetTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58(signature),
		&rpc.GetTransactionOpts{
			Encoding:                       solana.EncodingJSONParsed,
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: &cero,
		},
	)
	if err != nil {
		slog.Error(color.New(color.BgRed).SprintFunc()(fmt.Sprintf("Error fetching transaction: %v", err)))
		return
	}

	if tx.Meta != nil {
		for _, balance := range tx.Meta.PostTokenBalances {
			if balance.Owner.String() == "5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1" && balance.Mint.String() != "So11111111111111111111111111111111111111112" {
				slog.Info(color.New(color.BgYellow).SprintFunc()("========== New Token Found =========="))
				slog.Info(color.New(color.BgYellow).SprintFunc()(fmt.Sprintf("Mint Address: %s", balance.Mint)))
				slog.Info(color.New(color.BgYellow).SprintFunc()("====================================="))
			}
		}
	}
}
