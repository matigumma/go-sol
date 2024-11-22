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

	rpcClient := rpc.New(rpc.MainNetBeta_RPC)
	getTransactionDetails(rpcClient, signature)
}

func getTransactionDetails(rpcClient *rpc.Client, signature solana.Signature) {
	cero := uint64(0) // :/

	{
		// slog.Info(color.New(color.BgHiBlue).SprintFunc()("Fetching EncodingJSON transaction..."))
		// txJson, err := rpcClient.GetTransaction(
		// 	context.TODO(),
		// 	signature,
		// 	&rpc.GetTransactionOpts{
		// 		Encoding:                       solana.EncodingJSONParsed,
		// 		Commitment:                     rpc.CommitmentConfirmed,
		// 		MaxSupportedTransactionVersion: &cero,
		// 	},
		// )
		// if err != nil {
		// 	slog.Error(color.New(color.BgBlack, color.FgRed).SprintFunc()(fmt.Sprintf("Error fetching transaction: %v", err)))
		// }

		// if txJson != nil {
		// 	spew.Dump(txJson)
		// 	spew.Dump(txJson.Transaction.GetTransaction())
		// }

		// if txJson.Meta != nil {
		// 	slog.Info(color.New(color.BgHiBlue).SprintFunc()(fmt.Sprintf("Transaction details: %+v", txJson)))
		// 	for _, balance := range txJson.Meta.PostTokenBalances {
		// 		if balance.Owner.String() == "5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1" && balance.Mint.String() != "So11111111111111111111111111111111111111112" {
		// 			slog.Info(color.New(color.BgHiYellow).SprintFunc()("========== New Token Found =========="))
		// 			slog.Info(color.New(color.BgHiYellow).SprintFunc()(fmt.Sprintf("Mint Address: %s", balance.Mint)))
		// 			slog.Info(color.New(color.BgHiYellow).SprintFunc()("====================================="))
		// 		}
		// 	}
		// }
	}
	{
		slog.Info(color.New(color.BgHiBlue).SprintFunc()("Fetching EncodingBase58 transaction..."))
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
			slog.Error(color.New(color.BgBlack, color.FgRed).SprintFunc()(fmt.Sprintf("Error fetching EncodingBase58 transaction: %v", err)))
		}

		if tx58 != nil {
			// spew.Dump(tx58)
			// spew.Dump(tx58.Transaction.GetBinary())
			// if txJson.Meta != nil {
			for _, balance := range tx58.Meta.PostTokenBalances {
				if balance.Owner.String() == "5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1" && balance.Mint.String() != "So11111111111111111111111111111111111111112" {
					slog.Info(color.New(color.BgHiYellow).SprintFunc()("========== New Token Found =========="))
					slog.Info(color.New(color.BgHiYellow).SprintFunc()(fmt.Sprintf("Mint Address: %s", balance.Mint)))
					slog.Info(color.New(color.BgHiYellow).SprintFunc()("====================================="))
				}
			}
		}
	}
	// {
	// 	slog.Info(color.New(color.BgHiBlue).SprintFunc()("Fetching EncodingBase64 transaction..."))
	// 	tx64, err := rpcClient.GetTransaction(
	// 		context.TODO(),
	// 		signature,
	// 		&rpc.GetTransactionOpts{
	// 			Encoding:                       solana.EncodingBase64,
	// 			Commitment:                     rpc.CommitmentConfirmed,
	// 			MaxSupportedTransactionVersion: &cero,
	// 		},
	// 	)
	// 	if err != nil {
	// 		slog.Error(color.New(color.BgBlack, color.FgRed).SprintFunc()(fmt.Sprintf("Error fetching EncodingBase64 transaction: %v", err)))
	// 	}

	// 	if tx64 != nil {
	// 		spew.Dump(tx64)
	// 		spew.Dump(tx64.Transaction.GetBinary())
	// 	}
	// }

}
