package monitor

import (
	"context"
	"fmt"
	"log"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

func main() {
	client, err := ConnectToWebSocket()
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer client.Close()

	pubkey := "7YttLkHDoNj9wyDur5pM1ejNaAvT9X4eqaYcHQqtj2G5"
	if err := SubscribeToLogs(client, pubkey); err != nil {
		log.Fatalf("Failed to subscribe to logs: %v", err)
	}
}

func ConnectToWebSocket() (*ws.Client, error) {
	client, err := ws.Connect(context.Background(), rpc.MainNetBeta_WS)
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
	tx, err := rpcClient.GetTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58(signature),
		&rpc.GetTransactionOpts{
			Encoding:   solana.EncodingJSONParsed,
			Commitment: rpc.CommitmentConfirmed,
		},
	)
	if err != nil {
		fmt.Println("Error fetching transaction:", err)
		return
	}

	parsedTx := tx.Transaction.GetParsedTransaction()
	if parsedTx.Meta != nil {
		for _, balance := range parsedTx.Meta.PostTokenBalances {
			if balance.Owner == "5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1" && balance.Mint != "So11111111111111111111111111111111111111112" {
				fmt.Println("========== New Token Found ==========")
				fmt.Println("Mint Address:", balance.Mint)
				fmt.Println("=====================================")
			}
		}
	}
}
