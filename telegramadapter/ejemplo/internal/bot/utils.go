package bot

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

func (b *Bot) Answer(user *tg.User) *message.RequestBuilder {
	return b.Sender.To(user.AsInputPeer())
}

func (b *Bot) SetAnswerCallback(ctx context.Context, text string, queryID int64) error {
	_, err := b.API().MessagesSetBotCallbackAnswer(ctx, &tg.MessagesSetBotCallbackAnswerRequest{
		QueryID: queryID,
		Message: text,
	})
	return err
}

func (b *Bot) DeleteMessage(ctx context.Context, id int) error {
	_, err := b.API().MessagesDeleteMessages(ctx, &tg.MessagesDeleteMessagesRequest{
		Revoke: true,
		ID:     []int{id},
	})
	return err
}

func GetChannelByName(api *tg.Client, sender *message.Sender, ctx context.Context, name string) (*tg.Channel, error) {
	otherPeer, err := sender.Resolve(name).AsInputPeer(ctx)
	if err != nil {
		return nil, err
	}

	switch otherPeer.(type) {
	case *tg.InputPeerChannel:
		inputChannel, ok := peer.ToInputChannel(otherPeer)
		if !ok {
			return nil, fmt.Errorf("cannot cast to ToInputChannel")
		}

		/* maybe use ChannelsGetFullChannel ? */
		channels, err := api.ChannelsGetChannels(ctx, []tg.InputChannelClass{inputChannel})
		if err != nil {
			return nil, err
		}

		chats := channels.GetChats()
		if len(chats) == 0 {
			return nil, fmt.Errorf("getChats return empty slice")
		}

		return chats[0].(*tg.Channel), nil
	//case *tg.InputPeerChat:
	//api.MessagesGetFullChat(ctx)
	default:
		return nil, fmt.Errorf("it isn't channel or chat")
	}
}
