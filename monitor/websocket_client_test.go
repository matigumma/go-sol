package monitor_test

import (
	"context"
	"fmt"
	"gosol/monitor"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/stretchr/testify/assert"
)

func TestWebSocketMessageFlow(t *testing.T) {
	// Crear un canal para los logs
	logCh := make(chan *ws.LogResult, 1)
	statusCh := make(chan monitor.StatusMessage, 1)

	done := make(chan struct{})

	// Crear una instancia de App con el canal de logs
	app := &monitor.App{
		LogCh:         logCh,
		StatusUpdates: statusCh,
		Ctx:           context.Background(),
	}

	// Simular un mensaje de log
	expectedSignature := solana.MustSignatureFromBase58("g5Z5g5Z5g5Z5g5Z5g5Z5g5Z5g5Z5g5Z5g5Z5g5Z5g5Z5g5Z5g5Z5g5Z5g5Z5g5Z5g5Z5g5Z5g5Z5g5Z5g5Z5g5Z") // Ensure this is a valid 64-byte signature
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

	// Iniciar una goroutine para simular el procesamiento de logs en app
	go func() {
		select {
		case logMsg := <-app.LogCh:
			// Aquí puedes verificar que el mensaje se procesa correctamente
			fmt.Printf("Received log message: %v\n", logMsg)
			assert.Equal(t, expectedSignature, logMsg.Value.Signature)
			close(done)
		case <-time.After(1 * time.Second):
			t.Fatal("No se recibió el mensaje de log a tiempo")
		}
	}()

	// Enviar el mensaje simulado al canal de logs
	logCh <- logMsg

	<-done

	// Esperar un momento para que el mensaje sea procesado
	time.Sleep(100 * time.Millisecond)
}
