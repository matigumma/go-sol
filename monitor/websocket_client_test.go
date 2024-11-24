package monitor_test

import (
	"context"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/stretchr/testify/assert"
	"monitor"
)

func TestWebSocketMessageFlow(t *testing.T) {
	// Crear un canal para los logs
	logCh := make(chan *ws.LogResult, 1)
	statusCh := make(chan monitor.StatusMessage, 1)

	// Crear una instancia de App con el canal de logs
	app := &monitor.App{
		logCh:         logCh,
		StatusUpdates: statusCh,
		Ctx:           context.Background(),
	}

	// Simular un mensaje de log
	expectedSignature := "test-signature"
	logMsg := &ws.LogResult{
		Value: ws.LogValue{
			Signature: expectedSignature,
		},
	}

	// Iniciar una goroutine para simular el procesamiento de logs en app
	go func() {
		select {
		case logMsg := <-app.logCh:
			// Aquí puedes verificar que el mensaje se procesa correctamente
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
