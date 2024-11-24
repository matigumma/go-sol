package bot

import (
	"tg_reader_bot/internal/events"
)

func (b *Bot) handlePrivateMessage(msg events.MsgContext) error {
	err := b.Dispatch(msg.GetText(), msg)
	if err != nil {
		return err
	}

	if msg.UserData == nil {
		return nil
	}

	return b.stateHandler(msg)
}

func (b *Bot) handleChannelMessage(msg events.MsgContext) error {
	b.ParseIncomingMessage(msg)
	return nil
}

func (b *Bot) handleGroupChatMessage(msg events.MsgContext) error {
	b.ParseIncomingMessage(msg)
	return nil
}
