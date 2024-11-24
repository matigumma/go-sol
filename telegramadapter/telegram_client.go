package telegramadapter

import (
	"context"
	"log"
	"os"
	"regexp"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/joho/godotenv"
	"monitor"
)

func StartTelegramClient() {
	// Cargar variables de entorno
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiID := os.Getenv("API_ID")
	apiHash := os.Getenv("API_HASH")
	channelID := os.Getenv("TELEGRAM_CHANNEL_ID")

	// Crear cliente de Telegram
	client := telegram.NewClient(apiID, apiHash, telegram.Options{})

	// Conectar al cliente
	if err := client.Run(context.Background(), func(ctx context.Context) error {
		// Obtener el canal
		channel, err := getChannel(ctx, client, channelID)
		if err != nil {
			return err
		}

		// Configurar manejador de eventos
		client.OnNewMessage(func(ctx context.Context, msg *tg.Message) error {
			if msg.PeerID.ChannelID == channel.ID {
				processMessage(msg)
			}
			return nil
		})

		// Mantener la ejecución
		<-ctx.Done()
		return nil
	}); err != nil {
		log.Fatal(err)
	}
}

func getChannel(ctx context.Context, client *telegram.Client, channelID string) (*tg.Channel, error) {
	// Implementar lógica para obtener el canal usando el ID
	return nil, nil
}

func processMessage(msg *tg.Message) {
	// Filtrar mensajes que contienen "Platform: Raydium"
	if containsPlatformKeyword(msg.Message) {
		// Extraer dirección del token
		token := extractToken(msg.Message)
		if token != "" {
			// Enviar token al StateManager
			monitor.StateManager.AddMint(token)
		}
	}
}

func containsPlatformKeyword(message string) bool {
	// Obtener el valor de la variable de entorno PLATFORM_KEYWORD
	platformKeyword := os.Getenv("PLATFORM_KEYWORD")
	if platformKeyword == "" {
		// Valor por defecto si la variable de entorno no está configurada
		platformKeyword = "Platform: Raydium"
	}

	// Verificar si el mensaje contiene el valor de platformKeyword
	return regexp.MustCompile(regexp.QuoteMeta(platformKeyword)).MatchString(message)
}

func extractToken(message string) string {
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
