package wsmanager

import (
	"context"
	"fmt"
	"gosol/types"
	"gosol/monitor"
	"time"

	"github.com/gagliardetto/solana-go/rpc/ws"
)

type WSManager struct {
	client          *ws.Client
	websocketURL    string
	apiKey          string
	statusUpdates   chan<- monitor.StatusMessage
	tokenUpdates    chan<- []types.TokenInfo
}

func NewWSManager(websocketURL, apiKey string, statusUpdates chan<- monitor.StatusMessage, tokenUpdates chan<- []types.TokenInfo) *WSManager {
	return &WSManager{
		websocketURL:  websocketURL,
		apiKey:        apiKey,
		statusUpdates: statusUpdates,
		tokenUpdates:  tokenUpdates,
	}
}

func (w *WSManager) Connect() error {
	client, err := ws.Connect(context.Background(), w.websocketURL+w.apiKey)
	if err != nil {
		w.statusUpdates <- monitor.StatusMessage{Level: monitor.ERR, Message: fmt.Sprintf("Failed to connect to WebSocket: %v", err)}
		return err
	}
	w.client = client
	return nil
}

func (w *WSManager) SubscribeToLogs(pubkey string) error {
	if w.client == nil {
		return fmt.Errorf("WebSocket client is not connected")
	}

	sub, err := w.client.LogsSubscribeMentions(
		solana.MustPublicKeyFromBase58(pubkey),
		rpc.CommitmentConfirmed,
	)
	if err != nil {
		w.statusUpdates <- monitor.StatusMessage{Level: monitor.ERR, Message: fmt.Sprintf("Failed to subscribe to logs: %v", err)}
		return err
	}

	go func() {
		defer sub.Unsubscribe()
		w.statusUpdates <- monitor.StatusMessage{Level: monitor.WARN, Message: "Start monitoring..."}
		for {
			msg, err := sub.Recv(context.Background())
			if err != nil {
				w.statusUpdates <- monitor.StatusMessage{Level: monitor.ERR, Message: fmt.Sprintf("Error receiving log message: %v", err)}
				return
			}
			// Procesar el mensaje de log
			w.processLogMessage(msg)
		}
	}()
	return nil
}

func (w *WSManager) processLogMessage(msg *ws.LogResult) {
	// Implementa la lógica para procesar el mensaje de log
	// Por ejemplo, extraer la firma de la transacción y enviar para obtener detalles
	signature := msg.Value.Signature
	w.statusUpdates <- monitor.StatusMessage{Level: monitor.INFO, Message: fmt.Sprintf("Transaction Signature: %s", signature)}

	// Enviar la firma al procesador de logs o al manejador de estados
	// Puedes utilizar canales adicionales si es necesario
}

func (w *WSManager) Close() {
	if w.client != nil {
		w.client.Close()
	}
}
