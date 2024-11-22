package main

import (
	"context"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

// connectToSolana establece una conexión con la red de Solana.
func connectToSolana() *rpc.Client {
	client := rpc.New(rpc.MainNetBeta_RPC)
	return client
}

// subscribeToTransactions se suscribe a eventos de transacciones en la red de Solana.
func subscribeToTransactions() {
	client, err := ws.Connect(context.Background(), rpc.MainNetBeta_WS)
	if err != nil {
		panic(err)
	}

	sub, err := client.ProgramSubscribe(
		solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"), // Ejemplo de programa
		rpc.CommitmentRecent,
	)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	for {
		msg, err := sub.Recv(context.Background())
		if err != nil {
			panic(err)
		}
		// Procesa el mensaje de transacción
		fmt.Println("Transacción recibida:", msg)
	}
}

func main() {
	client := connectToSolana()
	fmt.Println("Conectado a Solana:", client)

	subscribeToTransactions()
}
