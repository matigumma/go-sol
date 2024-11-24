package bot

import (
	"context"
	"fmt"
	"math"
	"sort"
	"tg_reader_bot/internal/cache"
	"tg_reader_bot/internal/protobufs"

	"github.com/gotd/td/tg"
	"google.golang.org/protobuf/proto"
)

type menuInfo struct {
	page     int
	maxPages int
}

func (b *Bot) callbackAddNewPeer(btn buttonContext) error {
	rows := []tg.KeyboardButtonRow{CreateBackButton("❌ Отмена", protobufs.MessageID_MainPage, nil)}
	btn.UserData.State = cache.WaitingPeerName
	_, err := b.API().MessagesEditMessage(btn.Ctx, &tg.MessagesEditMessageRequest{
		Peer:        &tg.InputPeerUser{UserID: btn.Update.UserID},
		ID:          btn.UserData.ActiveMessageID,
		ReplyMarkup: &tg.ReplyInlineMarkup{Rows: rows},
		Message:     "🔗 Введите в чат ссылку/айди имя чата/группы.",
	})
	return err
}

func (b *Bot) showMyPeers(ctx context.Context, userCache *cache.UserData, QueryID int64, page int, sendNewMessage bool) error {
	if len(userCache.Peers) == 0 {
		return b.SetAnswerCallback(ctx, "📄 Список каналов пуст", QueryID)
	}

	var rows []tg.KeyboardButtonRow

	pagePeers, menuInfo := buildPage(page, userCache.Peers)
	rows = append(rows, buildMenuHeader(menuInfo))

	for _, id := range pagePeers {
		peer := userCache.Peers[id]
		rows = append(rows,
			CreateRowButton(
				peer.Title,
				protobufs.MessageID_PeerInfo,
				&protobufs.ButtonPeerInfo{PeerPage: int32(page), PeerId: peer.TelegramID, CurrentPage: 0},
			),
		)
	}

	rows = append(rows,
		buildNavigation(menuInfo, protobufs.MessageID_MyPeers, 0),
		CreateBackButton("↩️ Назад", protobufs.MessageID_MainPage, nil),
	)

	messageText := "💬 Ваши отслеживаемые каналы, нажмите, чтобы настроить.\n "
	if !sendNewMessage {
		_, err := b.API().MessagesEditMessage(ctx, &tg.MessagesEditMessageRequest{
			Peer:        &tg.InputPeerUser{UserID: userCache.TelegramID},
			ID:          userCache.ActiveMessageID,
			ReplyMarkup: &tg.ReplyInlineMarkup{Rows: rows},
			Message:     messageText,
		})
		return err
	} else {
		b.DeleteMessage(ctx, userCache.ActiveMessageID)
		_, err := b.Sender.To(&tg.InputPeerUser{UserID: userCache.TelegramID}).Markup(&tg.ReplyInlineMarkup{Rows: rows}).Text(ctx, messageText)
		return err
	}
}

func (b *Bot) callbackMyPeers(btn buttonContext) error {
	var message protobufs.ButtonMyPeers
	proto.Unmarshal(btn.Data, &message)

	return b.showMyPeers(btn.Ctx, btn.UserData, btn.Update.QueryID, int(message.CurrentPage), false)
}

func (b *Bot) showPeerInfo(ctx context.Context, peerID int64, tgUser *tg.User, keywordsPage int, peerPage int32, user *cache.UserData, sendNewMessage bool) error {
	peer, ok := user.Peers[peerID]
	if !ok {
		b.Answer(tgUser).Textf(ctx, "🛑 Ошибка при поиске канала.")
		return nil
	}

	user.ActivePeerID = peerID

	rows := []tg.KeyboardButtonRow{
		CreateRowButton(
			"📝 Новое ключевое слово",
			protobufs.MessageID_AddNewKeyWord,
			&protobufs.ButtonPeerInfo{PeerId: peerID},
		),
	}

	allkeywords := peer.GetUserKeyWords(user.GetID())

	pageKeywords, menuInfo := buildPage(keywordsPage, allkeywords)
	rows = append(rows, buildMenuHeader(menuInfo))

	for _, id := range pageKeywords {
		rows = append(rows,
			CreateRowButton(
				allkeywords[id],
				protobufs.MessageID_RemoveKeyWord,
				&protobufs.ButtonRemoveKeyWord{
					KeywordId: id,
					PeerInfo:  &protobufs.ButtonPeerInfo{PeerPage: peerPage, PeerId: peerID, CurrentPage: int32(menuInfo.page)},
				},
			),
		)
	}

	rows = append(rows,
		buildNavigation(menuInfo, protobufs.MessageID_PeerInfo, peerID),
		CreateButtonRow("🗑️ Удалить канал", protobufs.MessageID_RemovePeer, &protobufs.ButtonPeerInfo{PeerId: peerID}),
		CreateBackButton("↩️ Назад", protobufs.MessageID_MyPeers, &protobufs.ButtonMyPeers{CurrentPage: peerPage}),
		CreateBackButton("⤴️ На главную", protobufs.MessageID_MainPage, nil),
	)

	createStr := peer.CreatedAt.Format("2006-01-02 15:04:05")
	updateStr := peer.UpdatedAt.Format("2006-01-02 15:04:05")

	messageText := fmt.Sprintf(`
💬Канал: %s
📊Ключевых слов: %d
🗓️Добавлен: %s
🗓️Дата последнего обновления: %s
🗑️Нажмите на слово, чтобы его удалить.`, peer.Title, peer.GetUserKeyWordsCount(user.GetID()), createStr, updateStr)

	if !sendNewMessage {
		_, err := b.API().MessagesEditMessage(ctx, &tg.MessagesEditMessageRequest{
			Peer:        &tg.InputPeerUser{UserID: user.GetID()},
			ID:          user.ActiveMessageID,
			ReplyMarkup: &tg.ReplyInlineMarkup{Rows: rows},
			Message:     messageText,
		})
		return err
	} else {
		b.DeleteMessage(ctx, user.ActiveMessageID)
		_, err := b.Sender.To(&tg.InputPeerUser{UserID: user.GetID()}).Markup(&tg.ReplyInlineMarkup{Rows: rows}).Text(ctx, messageText)
		return err
	}
}

func (b *Bot) callbackPeerInfo(btn buttonContext) error {
	btn.UserData.State = cache.StateNone

	var message protobufs.ButtonPeerInfo
	proto.Unmarshal(btn.Data, &message)
	return b.showPeerInfo(btn.Ctx, message.PeerId, btn.User, int(message.CurrentPage), message.PeerPage, btn.UserData, false)
}

func (b *Bot) callbackBack(btn buttonContext) error {
	message := protobufs.ButtonMenuBack{}
	proto.Unmarshal(btn.Data, &message)

	if message.Msg != nil {
		btn.Data = message.Msg
	}

	return b.btnCallbacks[message.Newmenu](btn)
}

func (b *Bot) showMainPage(ctx context.Context, user *tg.User, userCache *cache.UserData, sendNewMessage bool) error {
	peersCount := 0
	if userCache != nil {
		userCache.State = cache.StateNone
		peersCount = len(userCache.Peers)
	}

	messageText := fmt.Sprintf(`
✨Добро пожаловать: %s %s ✨
💬Каналов отслеживается: %d 💬
`, user.FirstName, user.LastName, peersCount)

	if !sendNewMessage {
		_, err := b.API().MessagesEditMessage(ctx, &tg.MessagesEditMessageRequest{
			Peer:        &tg.InputPeerUser{UserID: user.ID},
			ID:          userCache.ActiveMessageID,
			ReplyMarkup: buildInitalMenu(),
			Message:     messageText,
		})
		return err
	} else {
		if userCache != nil {
			b.DeleteMessage(ctx, userCache.ActiveMessageID)
		}
		_, err := b.Sender.To(&tg.InputPeerUser{UserID: user.GetID()}).Markup(buildInitalMenu()).Text(ctx, messageText)
		return err
	}
}

func (b *Bot) callbackMainPage(btn buttonContext) error {
	return b.showMainPage(btn.Ctx, btn.User, btn.UserData, false)
}

func (b *Bot) callbackAddNewKeyWord(btn buttonContext) error {
	var message protobufs.ButtonPeerInfo
	proto.Unmarshal(btn.Data, &message)

	peer, ok := btn.UserData.Peers[message.PeerId]
	if !ok {
		b.Answer(btn.User).Text(btn.Ctx, "🛑 Ошибка при чтении ключевых слов.")
		return nil
	}

	btn.UserData.State = cache.WaitingKeyWord

	rows := []tg.KeyboardButtonRow{CreateBackButton("↩️ Назад", protobufs.MessageID_PeerInfo, &message)}
	_, err := b.API().MessagesEditMessage(btn.Ctx, &tg.MessagesEditMessageRequest{
		Peer:        &tg.InputPeerUser{UserID: btn.Update.UserID},
		ID:          btn.UserData.ActiveMessageID,
		ReplyMarkup: &tg.ReplyInlineMarkup{Rows: rows},
		Message:     fmt.Sprintf("💬Канал: %s \n✍Ввод ключевых слов, каждое новым сообщеним.", peer.Title),
	})

	return err
}

func (b *Bot) callbackRemoveKeyWord(btn buttonContext) error {
	var message protobufs.ButtonRemoveKeyWord
	proto.Unmarshal(btn.Data, &message)

	createNewMenu := false
	peer, ok := btn.UserData.Peers[message.PeerInfo.PeerId]
	if !ok {
		b.Answer(btn.User).Text(btn.Ctx, "🛑 Ошибка при удалении.")
		createNewMenu = true
	}

	err := peer.RemoveKeyword(btn.UserData.GetID(), message.KeywordId)
	if err != nil {
		b.Answer(btn.User).Text(btn.Ctx, "🛑 Ошибка при удалении.")
		createNewMenu = true
	}

	return b.showPeerInfo(btn.Ctx, message.PeerInfo.PeerId, btn.User, int(message.PeerInfo.CurrentPage), message.PeerInfo.PeerPage, btn.UserData, createNewMenu)
}

func (b *Bot) callbackRemovePeer(btn buttonContext) error {
	var message protobufs.ButtonPeerInfo
	proto.Unmarshal(btn.Data, &message)

	peer, ok := btn.UserData.Peers[message.PeerId]
	if !ok {
		b.Answer(btn.User).Text(btn.Ctx, "🛑 Ошибка при удалении.")
	}

	if b.peersCache.RemovePeerFromUser(btn.UserData, peer) != nil {
		b.Answer(btn.User).Textf(btn.Ctx, "🛑 Ошибка при удалении канала %s.", peer.Title)
	}

	b.Answer(btn.User).Textf(btn.Ctx, "✅ %s был удален.", peer.Title)

	if len(btn.UserData.Peers) == 0 {
		b.Answer(btn.User).Text(btn.Ctx, "⚠️ У вас нет отслеживаемых каналов, вы перемещены в главное меню.")
		return b.showMainPage(btn.Ctx, btn.User, btn.UserData, true)
	} else {
		return b.showMyPeers(btn.Ctx, btn.UserData, 0, 0, true)
	}
}

func (b *Bot) callbackSpaceButton(btn buttonContext) error {
	texts := []string{
		"Куда вы нажали?", "Больше не нажимайте на кнопку, вы сломаете бота!", "Упс, что-то пошло не так!",
		"Опять?!", "Вы серьезно?", "Я вас предупреждал!", "О, это становится интересно...",
		"Теперь это просто забавно!", "Мне нравится ваша настойчивость!", "Вы действительно упорны!", "Продолжайте, не останавливайтесь!",
	}

	textIndex := (btn.UserData.SecretButtonClicks / 3) % len(texts)
	btn.UserData.SecretButtonClicks += 1

	return b.SetAnswerCallback(btn.Ctx, texts[textIndex], btn.Update.QueryID)
}

func buildMenuHeader(menu menuInfo) tg.KeyboardButtonRow {
	if menu.page == -1 {
		return tg.KeyboardButtonRow{}
	}
	currentPage := fmt.Sprintf("📄 Страница %d/%d", menu.page+1, menu.maxPages)
	return CreateButtonRow(currentPage, protobufs.MessageID_Spacer, nil)
}

func buildNavigation(menu menuInfo, messageID protobufs.MessageID, peerID int64) tg.KeyboardButtonRow {
	var buttons []tg.KeyboardButtonClass

	if menu.page > 0 {
		prevPage := int32(menu.page - 1)
		if peerID != 0 {
			buttons = append(buttons, CreateButton("⬅️", messageID, &protobufs.ButtonPeerInfo{PeerId: peerID, CurrentPage: prevPage}))
		} else {
			buttons = append(buttons, CreateButton("⬅️", messageID, &protobufs.ButtonMyPeers{CurrentPage: prevPage}))
		}
	}

	if menu.page < menu.maxPages-1 {
		nextPage := int32(menu.page + 1)
		if peerID != 0 {
			buttons = append(buttons, CreateButton("➡️", messageID, &protobufs.ButtonPeerInfo{PeerId: peerID, CurrentPage: nextPage}))
		} else {
			buttons = append(buttons, CreateButton("➡️", messageID, &protobufs.ButtonMyPeers{CurrentPage: nextPage}))
		}
	}

	return tg.KeyboardButtonRow{
		Buttons: buttons,
	}
}

func buildPage[V any](curpage int, data map[int64]V) ([]int64, menuInfo) {
	maxPages := 0
	pageSize := 5

	maxPages = int(math.Ceil(float64(len(data)) / float64(pageSize)))

	if curpage < 0 {
		curpage = 0
	} else if curpage >= maxPages {
		curpage = maxPages - 1
	}

	startIndex := curpage * pageSize
	endIndex := startIndex + pageSize

	keys := make([]int64, 0, len(data))

	for k := range data {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	values := make([]int64, 0, pageSize)

	i := 0
	for _, id := range keys {
		if i >= startIndex && i < endIndex {
			values = append(values, id)
		}
		i++
	}

	return values, menuInfo{page: curpage, maxPages: maxPages}
}

func (b *Bot) registerQueryCallbacks() {
	b.btnCallbacks[protobufs.MessageID_AddNewPeer] = b.callbackAddNewPeer
	b.btnCallbacks[protobufs.MessageID_MyPeers] = b.callbackMyPeers
	b.btnCallbacks[protobufs.MessageID_AddNewKeyWord] = b.callbackAddNewKeyWord
	b.btnCallbacks[protobufs.MessageID_RemoveKeyWord] = b.callbackRemoveKeyWord
	b.btnCallbacks[protobufs.MessageID_Back] = b.callbackBack
	b.btnCallbacks[protobufs.MessageID_MainPage] = b.callbackMainPage
	b.btnCallbacks[protobufs.MessageID_PeerInfo] = b.callbackPeerInfo
	b.btnCallbacks[protobufs.MessageID_RemovePeer] = b.callbackRemovePeer
	b.btnCallbacks[protobufs.MessageID_Spacer] = b.callbackSpaceButton
}
