package monitor_test

import (
	"context"
	"gosol/monitor"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/stretchr/testify/assert"
)

func TestLogChannelMessageFlow(t *testing.T) {
	// Crear un canal para los logs
	logCh := make(chan *ws.LogResult, 1)

	// Crear una instancia de WebSocketClient con el canal de logs
	wsClient := &monitor.WebSocketClient{
		LogCh: logCh,
	}

	// Simular un mensaje de log
	expectedSignature := solana.MustSignatureFromBase58("test-signature")
	logMsg := &ws.LogResult{
		Value: struct {
			Signature solana.Signature `json:"signature"`
			Err       interface{}      `json:"err"`
			Logs      []string         `json:"logs"`
		}{
			Signature: solana.Signature(expectedSignature),
			Err:       nil,
			Logs:      nil,
		},
	}

	// Iniciar una goroutine para simular el procesamiento de logs en wsClient
	go func() {
		select {
		case logMsg := <-wsClient.LogCh:
			// Verificar que el mensaje se procesa correctamente
			assert.Equal(t, expectedSignature, logMsg.Value.Signature)
		case <-time.After(1 * time.Second):
			t.Fatal("No se recibió el mensaje de log a tiempo")
		}
	}()

	// Enviar el mensaje simulado al canal de logs
	logCh <- logMsg

	// Esperar un momento para que el mensaje sea procesado
	time.Sleep(100 * time.Millisecond)
}
