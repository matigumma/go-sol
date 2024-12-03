package telegramadapter

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tg "matu/gosol/gogram/telegram"
	"matu/gosol/monitor"

	"github.com/joho/godotenv"
)

type TelegramListener struct {
	monitor      *monitor.App
	botID        int64
	messageQueue chan *tg.NewMessage
	client       *tg.Client // Add client field to use the same instance
}

func NewTelegramListener(monitor *monitor.App, botID int64) *TelegramListener {
	return &TelegramListener{
		monitor:      monitor,
		botID:        botID,
		messageQueue: make(chan *tg.NewMessage, 100),
	}
}

func (t *TelegramListener) Run() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}

	appID := os.Getenv("API_ID")
	appIDInt, _ := strconv.Atoi(appID)
	appHash := os.Getenv("API_HASH")

	// Create the Telegram client
	client, err := tg.NewClient(tg.ClientConfig{
		AppID:         int32(appIDInt),
		AppHash:       appHash,
		LogLevel:      tg.LogInfo,
		MemorySession: true,
		StringSession: func() string {
			sessionData, err := os.ReadFile("session.session")
			if err != nil {
				return ""
			}
			return string(sessionData)
		}(),
	}, nil)
	if err != nil {
		log.Fatal(err)
	}

	t.client = client // Store the client instance

	client.On(tg.OnMessage, func(message *tg.NewMessage) error {
		t.messageQueue <- message
		return nil
	}, tg.Filter{
		Group:   false,
		FromBot: true,
		Chats:   []int64{t.botID},
	})

	go t.processMessageQueue()

	client.Start()
	client.Idle()
}

func (t *TelegramListener) SendMessage(chatID int64, text string) {
	_, err := t.client.SendMessage(chatID, text, nil) // Use the existing client instance
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (t *TelegramListener) processMessageQueue() {
	for {
		msg := <-t.messageQueue
		fmt.Printf("Received message: %+v\n", msg.Text())
		// Handle the message as needed
		time.Sleep(2 * time.Second) // Simulate processing time
	}
}
