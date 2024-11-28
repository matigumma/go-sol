package telegramadapter

import (
	"encoding/json"
	"fmt"
	"gosol/logger"
	"gosol/monitor"
	"log"
	"os"
	"regexp"
	"strconv"

	tg "matu/gogram/telegram"

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
		return
	}

	appID := os.Getenv("API_ID")
	appIDInt, _ := strconv.Atoi(appID)

	appHash := os.Getenv("API_HASH")
	// botToken := os.Getenv("BOT_TOKEN")
	phoneNum := os.Getenv("PHONE")
	fmt.Printf("Phone number: %s", phoneNum)

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

	logger := logger.NewLogger("gosol [telegram]", t.monitor.StatusUpdates)

	client, err := tg.NewClient(tg.ClientConfig{
		AppID:    int32(appIDInt), // https://my.telegram.org/auth?to=apps
		AppHash:  appHash,
		LogLevel: tg.LogDebug,
		// PublicKeys: rsaPublicKeys,
		MemorySession: true,
		StringSession: func() string {
			sessionData, err := os.ReadFile("session.session")
			if err != nil {
				// log.Fatal("Error reading session from file:", err)
				return ""
			}
			return string(sessionData)
		}(),
	}, logger)
	if err != nil {
		log.Fatal(err)
	}

	// client.Conn()

	// // Authenticate the client using the bot token
	// // This will send a code to the phone number if it is not already authenticated
	// // if err := client.LoginBot(botToken); err != nil {
	// if _, err := client.Login(phoneNum); err != nil {
	// 	panic(err)
	// }

	client.Start()

	sessionData := client.ExportSession()

	err = os.WriteFile("session.session", []byte(sessionData), 0644)
	if err != nil {
		errSess := fmt.Sprintln("Error writing session to file:", err)
		t.monitor.StatusUpdates <- monitor.StatusMessage{Level: monitor.ERR, Message: errSess}
	}

	// client.UpdatesGetState()

	client.On(tg.OnMessage, func(message *tg.NewMessage) error { // client.AddMessageHandler
		t.processMessage(message)

		return nil
	}, tg.Filter{
		Group: true,
		// Chats: []int64{227962}, // 227963 Raydium, 227963 pump fun thread {Message.ReplyTo.ReplyToMsgID}
		// Func: func(message *tg.NewMessage) bool {
		// 	// esto es para testear si esta leyendo mensajes de otro canal que no sea el de Raydium
		// 	if message.Message.Replies.ChannelID == int64(-1002109566555) {
		// 		return false
		// 	}
		// 	return true
		// },
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

		messageJSON, err := json.MarshalIndent(msg, "", "  ")
		if err == nil {
			// Append the message to a JSON file
			file, err := os.OpenFile("messages.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				errOpen := fmt.Sprintf("Error opening file: %v", err)
				t.monitor.StatusUpdates <- monitor.StatusMessage{Level: monitor.ERR, Message: errOpen}
			} else {
				defer file.Close()
				if _, err := file.WriteString(string(messageJSON) + "\n"); err != nil {
					erMsg := fmt.Sprintf("Error writing to file: %v", err)
					t.monitor.StatusUpdates <- monitor.StatusMessage{Level: monitor.ERR, Message: erMsg}
				}
			}
		}

		// Extraer dirección del token
		token := t.extractToken(msg.Message.Message)
		if token != "" {
			t.monitor.StatusUpdates <- monitor.StatusMessage{Level: monitor.INFO, Message: "✅ New Token: " + token}
			// log.Printf("New Token: %s", token)
			// Enviar token al StateManager
			t.monitor.StateManager.AddMint(token)
			t.monitor.ApiClient.FetchAndProcessReport(token)
		}
		// } else {
		// 	fmt.Print(".")
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
