package telegram

import (
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
)

type TGClient struct {
	Client *telegram.Client
	Sender *message.Sender
}

func InitTGClient(client *telegram.Client) *TGClient {
	return &TGClient{Client: client, Sender: message.NewSender(client.API())}
}
