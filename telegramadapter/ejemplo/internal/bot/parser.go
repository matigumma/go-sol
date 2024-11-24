package bot

import (
	"context"
	"fmt"
	"strings"
	"tg_reader_bot/internal/app"
	"tg_reader_bot/internal/cache"
	"tg_reader_bot/internal/events"
	"tg_reader_bot/internal/models"
	"time"

	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

func (bot *Bot) ParseChannels(ctx context.Context) {
	for {
		app := app.GetContainer()
		if app.Client != nil {
			db := app.Database

			tgclient := app.Client.Client
			cache := &bot.peersCache
			cache.Mutex.Lock()
			for _, peerInfo := range cache.Peers {
				history, err := tgclient.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
					Peer:  &tg.InputPeerChannel{ChannelID: peerInfo.TelegramID, AccessHash: peerInfo.AccessHash},
					Limit: 10,
					MinID: peerInfo.LastMsgID,
				})

				if err != nil {
					fmt.Println("Failed to MessagesGetHistory", err)
					continue
				}

				modifed, ok := history.AsModified()
				if !ok {
					fmt.Println("Failed to cast to AsModified")
					continue
				}

				messages := modifed.GetMessages()
				if len(messages) == 0 {
					continue
				}

				for _, message := range messages {
					tgmessage, ok := message.(*tg.Message)
					if !ok {
						continue
					}
					bot.FindUsersKeyWords(ctx, tgmessage, peerInfo)
				}

				/* a very crappy lib */
				var id int
				switch v := messages[0].(type) {
				case *tg.MessageEmpty:
					id = v.ID
				case *tg.Message:
					id = v.ID
				case *tg.MessageService:
					id = v.ID
				}

				peerInfo.LastMsgID = id
			}

			for _, peerInfo := range cache.Peers {
				db.Model(&models.Peer{ID: peerInfo.DatabaseID}).Updates(models.Peer{LastMessageID: peerInfo.LastMsgID})
				peerInfo.UpdatedAt = time.Now()
			}

			cache.Mutex.Unlock()

			time.Sleep(5 * time.Second)
		}
		time.Sleep(30 * time.Second)
	}
}

func (bot *Bot) ParseIncomingMessage(msg events.MsgContext) {
	cache := &bot.peersCache

	/* a very crappy lib */
	var peerID int64
	switch v := msg.Message.FromID.(type) {
	case *tg.PeerChat:
		peerID = v.ChatID
	case *tg.PeerChannel:
		peerID = v.ChannelID
	}

	peer, ok := cache.Peers[peerID]
	if !ok {
		return
	}

	bot.FindUsersKeyWords(msg.Ctx, msg.Message, peer)
}

func CaseInsensitiveContains(s, substr string) bool {
	s, substr = strings.ToUpper(s), strings.ToUpper(substr)
	return strings.Contains(s, substr)
}

func (bot *Bot) FindUsersKeyWords(ctx context.Context, message *tg.Message, peerInfo *cache.PeerData) {
	for userID, users := range peerInfo.UsersKeyWords {
		for _, keyword := range users.Keywords {
			if !CaseInsensitiveContains(message.Message, keyword) {
				continue
			}

			var fromID int64
			switch v := message.FromID.(type) {
			case *tg.PeerUser:
				fromID = v.UserID
			case *tg.PeerChat:
				fromID = v.ChatID
			case *tg.PeerChannel:
				fromID = v.ChannelID
			}

			entities := []styling.StyledTextOption{
				styling.Plain("📨 Найдено ключевое слово: "), styling.Bold(keyword), styling.Plain("\n"),
				styling.Plain("💬 Канал: "), styling.TextURL(peerInfo.Title, fmt.Sprintf("https://t.me/%s/", peerInfo.UserName)), styling.Plain("\n"),
				styling.TextURL("✉️ Сообщение в чате", fmt.Sprintf("https://t.me/%s/%d", peerInfo.UserName, message.ID)), styling.Plain("\n"),
				styling.MentionName("🧑🏻‍💻 Пользователь", &tg.InputUser{UserID: fromID}), styling.Plain("\n"),
				styling.Plain(fmt.Sprintf("📜 Текст: %s", message.Message)),
			}

			bot.Sender.To(&tg.InputPeerUser{UserID: userID}).NoWebpage().StyledText(ctx, entities...)
			break
		}
	}
}
