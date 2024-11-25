package telegramadapter

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"gosol/monitor"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type TelegramClient struct {
	monitor *monitor.App
}

func NewTelegramClient(monitor *monitor.App) *TelegramClient {
	fmt.Println("New Telegram Client model")
	return &TelegramClient{
		monitor: monitor,
	}
}

func (t *TelegramClient) Run() {
	fmt.Println("Running Telegram Client")
	// Cargar variables de entorno
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	apiID := os.Getenv("API_ID")
	apiIDInt, err := strconv.Atoi(apiID)
	if err != nil {
		log.Fatal("Error converting API_ID to int:", err)
	}
	apiHash := os.Getenv("API_HASH")
	tchannelID := os.Getenv("TELEGRAM_CHANNEL_ID")
	channelID, err := strconv.Atoi(tchannelID)
	if err != nil {
		log.Fatal("Error converting API_ID to int:", err)
	}

	// Crear cliente de Telegram
	client := telegram.NewClient(apiIDInt, apiHash, telegram.Options{})
	fmt.Println("Client created")

	// Conectar al cliente
	if err := client.Run(context.Background(), func(ctx context.Context) error {

		t.monitor.StatusUpdates <- monitor.StatusMessage{Level: monitor.INFO, Message: "Connected to Telegram"}

		dispatcher := tg.NewUpdateDispatcher()

		dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
			t.monitor.StatusUpdates <- monitor.StatusMessage{Level: monitor.INFO, Message: "New message received"}

			msg, ok := update.Message.(*tg.Message)
			if !ok {
				return nil
			}

			t.monitor.StatusUpdates <- monitor.StatusMessage{Level: monitor.INFO, Message: msg.Message[:10]}

			if msg.Replies.ChannelID == int64(channelID) {
				t.processMessage(msg)
			}

			return nil // Return nil if no error occurs
		})

		// Registrar el event handler para nuevos mensajes de canal
		dispatcher.OnNewChannelMessage(t.onNewChannelMessage)

		// Mantener la ejecución
		<-ctx.Done()
		return nil
	}); err != nil {
		log.Fatal(err)
	}
}

func (t *TelegramClient) processMessage(msg *tg.Message) {
	// Filtrar mensajes que contienen "Platform: Raydium || Pump Fun"
	if containsPlatformKeyword(msg.Message) {
		// Extraer dirección del token
		token := extractToken(msg.Message)
		if token != "" {
			t.monitor.StatusUpdates <- monitor.StatusMessage{Level: monitor.INFO, Message: "New Token Found: " + token}
			// Enviar token al StateManager
			t.monitor.StateManager.AddMint(token)
		}
	}
}

func containsPlatformKeyword(message string) bool {
	// Obtener el valor de la variable de entorno PLATFORM_KEYWORD
	platformKeyword := os.Getenv("PLATFORM_KEYWORD")
	if platformKeyword == "" {
		// Valor por defecto si la variable de entorno no está configurada
		platformKeyword = "Raydium"
	}

	// Asegurarse de que el valor de platformKeyword siempre comience con "Platform: "
	fullKeyword := "Platform: " + platformKeyword

	// Verificar si el mensaje contiene el valor de fullKeyword
	return regexp.MustCompile(regexp.QuoteMeta(fullKeyword)).MatchString(message)
}

func (t *TelegramClient) onNewChannelMessage(ctx context.Context, entities tg.Entities, update *tg.UpdateNewChannelMessage) error {
	t.monitor.StatusUpdates <- monitor.StatusMessage{Level: monitor.INFO, Message: "New channel message received"}

	msg, ok := update.Message.(*tg.Message)
	if !ok {
		return nil
	}

	// Procesar el mensaje del canal
	t.processMessage(msg)

	return nil
}
	// Verificar si el mensaje contiene "Platform: Raydium"
	if !containsPlatformKeyword(message) {
		return ""
	}

	// Dividir el mensaje en partes usando "Base: " como separador
	parts := regexp.MustCompile(`Base: `).Split(message, 2)
	if len(parts) > 1 {
		// Extraer dirección antes de "\nQuote:"
		addressParts := regexp.MustCompile(`\nQuote:`).Split(parts[1], 2)
		if len(addressParts) > 0 {
			address := addressParts[0]
			return address
		}
	}

	return ""
}
