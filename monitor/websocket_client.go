package monitor

import (
	"context"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

type WebSocketClient struct {
	client        *ws.Client
	websocketURL  string
	apiKey        string
	pubkey        string
	logCh         chan<- *ws.LogResult
	statusUpdates chan<- StatusMessage
}

func NewWebSocketClient(websocketURL, apiKey, pubkey string, logCh chan<- *ws.LogResult, statusUpdates chan<- StatusMessage) *WebSocketClient {
	return &WebSocketClient{
		websocketURL:  websocketURL,
		apiKey:        apiKey,
		pubkey:        pubkey,
		logCh:         logCh,
		statusUpdates: statusUpdates,
	}
}

func (wsc *WebSocketClient) Connect(ctx context.Context) error {
	client, err := ws.Connect(ctx, wsc.websocketURL+wsc.apiKey)
	if err != nil {
		wsc.updateStatus(fmt.Sprintf("Failed to connect to WebSocket: %v", err))
		return err
	}
	wsc.client = client
	return nil
}

func (wsc *WebSocketClient) Subscribe(ctx context.Context) error {
	program := solana.MustPublicKeyFromBase58(wsc.pubkey)

	sub, err := wsc.client.LogsSubscribeMentions(
		program,
		rpc.CommitmentConfirmed,
	)
	if err != nil {
		wsc.updateStatus(fmt.Sprintf("Failed to subscribe to logs: %v", err))
		return err
	}

	go func() {
		defer sub.Unsubscribe()
		wsc.updateStatus("Start monitoring...", INFO)
		for {
			msg, err := sub.Recv(ctx)
			if err != nil {
				wsc.updateStatus(fmt.Sprintf("WebSocket error: %v", err))
				return
			}
			wsc.logCh <- msg
		}
	}()

	return nil
}

func (wsc *WebSocketClient) Reconnect(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := wsc.Connect(ctx)
			if err != nil {
				wsc.updateStatus("Retrying connection in 5 seconds...")
				time.Sleep(5 * time.Second)
				continue
			}
			err = wsc.Subscribe(ctx)
			if err != nil {
				wsc.updateStatus("Retrying subscription in 5 seconds...")
				time.Sleep(5 * time.Second)
				continue
			}
			break
		}
	}
}

func (wsc *WebSocketClient) updateStatus(message string) {
	wsc.statusUpdates <- StatusMessage{Level: INFO, Message: message}
}
