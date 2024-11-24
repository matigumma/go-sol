package bot

import (
	"tg_reader_bot/internal/app"
	"tg_reader_bot/internal/cache"
	"tg_reader_bot/internal/events"
)

func (b *Bot) enterPeerName(msg events.MsgContext) error {
	client := app.GetClient()

	user := msg.UserData

	b.DeleteMessage(msg.Ctx, msg.Message.ID)

	peer, err := GetChannelByName(client.Client.API(), client.Sender, msg.Ctx, msg.GetText())
	if err != nil {
		b.Answer(msg.PeerUser).Text(msg.Ctx, "🛑 Ошибка при выполнении. Попробуйте позже.")
		return err
	}

	if user.HasPeerByID(peer.ID) {
		b.Answer(msg.PeerUser).NoWebpage().Textf(msg.Ctx, "🛑 %s уже добавлен.", peer.Title)
		return b.showPeerInfo(msg.Ctx, peer.ID, msg.PeerUser, 0, 0, user, true)
	}

	err = b.peersCache.AddPeerToUser(msg.UserData, peer)
	if err != nil {
		b.Answer(msg.PeerUser).Text(msg.Ctx, "🛑 Ошибка при выполнении. Попробуйте позже.")
		return err
	}

	_, err = b.Answer(msg.PeerUser).NoWebpage().Textf(msg.Ctx, "✅ %s успешно добавлен.", peer.Title)
	if err != nil {
		return err
	}

	return b.showPeerInfo(msg.Ctx, peer.ID, msg.PeerUser, 0, 0, user, true)
}

func (b *Bot) enterKeyWord(msg events.MsgContext) error {
	user := msg.UserData

	b.DeleteMessage(msg.Ctx, msg.Message.ID)

	peer := user.GetActivePeer()
	if peer == nil {
		b.Answer(msg.PeerUser).Textf(msg.Ctx, "🛑 Ошибка при получении канала.")
		return nil
	}

	err := peer.AddKeyword(user, msg.GetText())
	if err != nil {
		return err
	}

	return nil
}

func (b *Bot) stateHandler(msg events.MsgContext) error {
	switch msg.UserData.State {
	case cache.WaitingPeerName:
		return b.enterPeerName(msg)
	case cache.WaitingKeyWord:
		return b.enterKeyWord(msg)
	default:
		return nil
	}
}
