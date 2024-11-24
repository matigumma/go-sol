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
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := sub.Recv(ctx)
				if err != nil {
					wsc.updateStatus(fmt.Sprintf("WebSocket error: %v", err), ERR)
					return
				}
				wsc.updateStatus("Subscribe: Received log message", INFO)
				wsc.updateStatus("Sending log message to logCh", INFO)
				select {
				case wsc.logCh <- msg:
					wsc.updateStatus("Log message sent to logCh", INFO)
				case <-ctx.Done():
					wsc.updateStatus("Context done before sending log message", WARN)
					return
				}
			}
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
			if err := wsc.Connect(ctx); err != nil {
				wsc.updateStatus(fmt.Sprintf("Connection failed: %v. Retrying...", err), ERR)
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			}

			if err := wsc.Subscribe(ctx); err != nil {
				wsc.updateStatus(fmt.Sprintf("Subscription failed: %v. Retrying...", err), ERR)
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			}

			wsc.updateStatus("Successfully reconnected and subscribed.", INFO)
			backoff = 1 * time.Second // Reset backoff after a successful connection
			return
		}
	}
}

func (wsc *WebSocketClient) updateStatus(message string, level LogLevel) {
	wsc.statusUpdates <- StatusMessage{Level: level, Message: message}
}
