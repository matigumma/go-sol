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
	logCh         chan<- *ws.LogResult
	statusUpdates chan<- StatusMessage
}

func NewWebSocketClient(logCh chan<- *ws.LogResult, statusUpdates chan<- StatusMessage) *WebSocketClient {
	return &WebSocketClient{
		logCh:         logCh,
		statusUpdates: statusUpdates,
	}
}

func (wsc *WebSocketClient) Connect(ctx context.Context) error {
	url := fmt.Sprintf("%s%s", websocketURL, apiKey)

	client, err := ws.Connect(ctx, url)
	if err != nil {
		wsc.updateStatus(fmt.Sprintf("Failed to connect to WebSocket: %v", err), ERR)
		return err
	}
	wsc.client = client
	return nil
}

func (wsc *WebSocketClient) Subscribe(ctx context.Context) error {
	program := solana.MustPublicKeyFromBase58(pubkey)

	sub, err := wsc.client.LogsSubscribeMentions(
		program,
		rpc.CommitmentConfirmed,
	)
	if err != nil {
		wsc.updateStatus(fmt.Sprintf("Failed to subscribe to logs: %v", err), ERR)
		return err
	}

	wsc.updateStatus("Start monitoring...", INFO)
	go func() {
		defer sub.Unsubscribe()
		for {
			msg, err := sub.Recv(ctx)
			if err != nil {
				wsc.updateStatus(fmt.Sprintf("WebSocket error: %v", err), ERR)
				return
			}
			wsc.logCh <- msg
		}
	}()

	return nil
}

func (wsc *WebSocketClient) Reconnect(ctx context.Context) {
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := wsc.Connect(ctx)
			if err != nil {
				wsc.updateStatus("Retrying connection...", ERR)
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			}
			err = wsc.Subscribe(ctx)
			if err != nil {
				wsc.updateStatus("Retrying subscription...", ERR)
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			}
			backoff = 1 * time.Second // Reset backoff after a successful connection
			break
		}
	}
}

func (wsc *WebSocketClient) updateStatus(message string, level LogLevel) {
	wsc.statusUpdates <- StatusMessage{Level: level, Message: message}
}
