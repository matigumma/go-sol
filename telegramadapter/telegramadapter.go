package telegramadapter

import (
	"encoding/json"
	"fmt"
	"gosol/monitor"
	"log"
	"os"
	"regexp"
	"strconv"

	tg "github.com/amarnathcjd/gogram/telegram"
	"github.com/joho/godotenv"
)

type TelegramClient struct {
	monitor         *monitor.App
	platformKeyword string
}

func NewTelegramClient(monitor *monitor.App) *TelegramClient {
	return &TelegramClient{
		monitor: monitor,
	}
}

func (t *TelegramClient) Run() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	appID := os.Getenv("API_ID")
	appIDInt, _ := strconv.Atoi(appID)

	appHash := os.Getenv("API_HASH")
	// botToken := os.Getenv("BOT_TOKEN")
	// phoneNum := os.Getenv("PHONE")

	// tchannelID := os.Getenv("TELEGRAM_CHANNEL_ID")
	// channelID, err := strconv.Atoi(tchannelID)
	// if err != nil {
	// 	log.Fatal("Error converting API_ID to int:" + err.Error())
	// }

	t.platformKeyword = os.Getenv("PLATFORM_KEYWORD")

	// rsapubkey := os.Getenv("TEST_RSA_PUB_KEY")

	// var rsaPublicKeys []*rsa.PublicKey
	// if rsapubkey != "" {
	// 	block, _ := pem.Decode([]byte(rsapubkey))
	// 	if block == nil {
	// 		log.Fatal("Failed to parse PEM block containing the public key")
	// 	}

	// 	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	// 	if err != nil {
	// 		log.Fatal("Failed to parse DER encoded public key: ", err)
	// 	}

	// 	switch pub := pub.(type) {
	// 	case *rsa.PublicKey:
	// 		rsaPublicKeys = append(rsaPublicKeys, pub)
	// 	default:
	// 		log.Fatal("Key is not a valid RSA public key")
	// 	}
	// }

	client, err := tg.NewClient(tg.ClientConfig{
		AppID:    int32(appIDInt), // https://my.telegram.org/auth?to=apps
		AppHash:  appHash,
		LogLevel: tg.LogInfo,
		// PublicKeys: rsaPublicKeys,
		MemorySession: true,
		StringSession: func() string {
			sessionData, err := os.ReadFile("session.session")
			if err != nil {
				log.Fatal("Error reading session from file:", err)
				return ""
			}
			return string(sessionData)
		}(),
	})
	if err != nil {
		log.Fatal(err)
	}

	// client.Conn()

	// Authenticate the client using the bot token
	// This will send a code to the phone number if it is not already authenticated
	// if err := client.LoginBot(botToken); err != nil {
	// if _, err := client.Login(phoneNum); err != nil {
	// 	panic(err)
	// }

	client.Start()

	sessionData := client.ExportSession()

	err = os.WriteFile("session.session", []byte(sessionData), 0644)
	if err != nil {
		log.Fatal("Error writing session to file:", err)
	}

	// client.UpdatesGetState()

	client.On(tg.OnMessage, func(message *tg.NewMessage) error { // client.AddMessageHandler
		log.Printf("Received message")

		messageJSON, err := json.MarshalIndent(message, "", "  ")
		if err != nil {
			log.Println("Error marshaling message to JSON:", err)
			return err
		}
		fmt.Printf("Received message: %s\n", messageJSON)

		os.Exit(0)

		return nil
	}, tg.Filter{
		Group: true,
		Func: func(message *tg.NewMessage) bool {
			if message.Message.Replies.ChannelID == int64(-1002109566555) {
				return false
			}
			return true
		},
		// Chats: []int64{-1002109566555},
	})

	// "Message": {
	// 		"PeerID": {
	// 			"ChannelID": 2109566555
	// 		},
	// },
	// "Peer": {
	// 		"ChannelID": 2109566555,
	// 		"AccessHash": -6807968663726781214
	// },

	// client Do anything
	client.Idle()
}

func (t *TelegramClient) processMessage(msg *tg.NewMessage) {
	// Filtrar mensajes que contienen "Platform: Raydium || Pump Fun"
	if t.containsPlatformKeyword(msg.Message.Message) {
		// Extraer dirección del token
		token := t.extractToken(msg.Message.Message)
		if token != "" {
			// t.monitor.StatusUpdates <- monitor.StatusMessage{Level: monitor.INFO, Message: "New Token Found: " + token}
			log.Printf("New Token: %s", token)
			// Enviar token al StateManager
			// t.monitor.StateManager.AddMint(token)
		}
	}
}

func (t *TelegramClient) containsPlatformKeyword(message string) bool {
	// Obtener el valor de la variable de entorno PLATFORM_KEYWORD

	if t.platformKeyword == "" {
		// Valor por defecto si la variable de entorno no está configurada
		t.platformKeyword = "Raydium"
	}

	// Asegurarse de que el valor de platformKeyword siempre comience con "Platform: "
	fullKeyword := "Platform: " + t.platformKeyword

	// Verificar si el mensaje contiene el valor de fullKeyword
	return regexp.MustCompile(regexp.QuoteMeta(fullKeyword)).MatchString(message)
}

func (t *TelegramClient) extractToken(message string) string {
	// Verificar si el mensaje contiene "Platform: Raydium"
	if !t.containsPlatformKeyword(message) {
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
