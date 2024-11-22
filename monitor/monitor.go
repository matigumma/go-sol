package monitor

import (
	"context"
	"fmt"
	"log"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

func Run() {
	client, err := ConnectToWebSocket()
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer client.Close()

	pubkey := "7YttLkHDoNj9wyDur5pM1ejNaAvT9X4eqaYcHQqtj2G5" // ray_fee_pubkey
	if err := SubscribeToLogs(client, pubkey); err != nil {
		log.Fatalf("Failed to subscribe to logs: %v", err)
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

	fmt.Println("Start monitoring...")
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
	fmt.Println("Transaction Signature:", signature)

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
		fmt.Println("Error fetching transaction:", err)
		return
	}

	if tx.Meta != nil {
		for _, balance := range tx.Meta.PostTokenBalances {
			if balance.Owner.String() == "5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1" && balance.Mint.String() != "So11111111111111111111111111111111111111112" {
				fmt.Println("========== New Token Found ==========")
				fmt.Println("Mint Address:", balance.Mint)
				fmt.Println("=====================================")
			}
		}
	}
}
