package monitor

import (
	"context"
	"fmt"
	"gosol/types"
	"sync"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type TransactionManager struct {
	rpcClient       *rpc.Client
	apiClient       *APIClient
	stateManager    *StateManager
	statusUpdates   chan<- StatusMessage
	tokenUpdates    chan<- []types.TokenInfo
	wg              sync.WaitGroup
	requestThrottle chan struct{}
}

func NewTransactionManager(apiClient *APIClient, stateManager *StateManager, statusUpdates chan<- StatusMessage, tokenUpdates chan<- []types.TokenInfo) *TransactionManager {
	rpcURL := "https://mainnet.helius-rpc.com/?api-key=" + apiKey
	return &TransactionManager{
		rpcClient:       rpc.New(rpcURL),
		apiClient:       apiClient,
		stateManager:    stateManager,
		statusUpdates:   statusUpdates,
		tokenUpdates:    tokenUpdates,
		requestThrottle: make(chan struct{}, 10), // Limitar a 10 consultas concurrentes
	}
}

func (tm *TransactionManager) HandleTransaction(signature solana.Signature) {
	tm.wg.Add(1)
	go func(sig solana.Signature) {
		defer tm.wg.Done()
		tm.fetchAndProcessTransaction(sig)
	}(signature)
}

func (tm *TransactionManager) fetchAndProcessTransaction(signature solana.Signature) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updateStatus := func(message string, level LogLevel) {
		tm.statusUpdates <- StatusMessage{Level: level, Message: message}
	}

	tx, err := tm.rpcClient.GetTransaction(
		ctx,
		signature,
		&rpc.GetTransactionOpts{
			Encoding:                       solana.EncodingBase58,
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: nil, // Usa el valor por defecto
		},
	)
	if err != nil {
		// updateStatus(fmt.Sprintf("Error fetching transaction %s", signature), ERR)
		return
	}

	if tx != nil {
		for _, balance := range tx.Meta.PostTokenBalances {
			if balance.Owner.String() == "5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1" && balance.Mint.String() != "So11111111111111111111111111111111111111112" {
				updateStatus(fmt.Sprintf("========== New Token Found: %s ==========", balance.Mint.String()), INFO)
				tm.stateManager.AddMint(balance.Mint.String())
				tm.apiClient.FetchAndProcessReport(balance.Mint.String())
			}
		}
	}
}

func (tm *TransactionManager) Wait() {
	tm.wg.Wait()
}
