package bot

import (
	"tg_reader_bot/internal/events"
)

func (b *Bot) startCommand(msg events.MsgContext) error {
	welcomeText := "🤖 Я бот для отслеживания сообщений в чатах и каналах.\n" +
		"⚙️ Ты можешь добавить необходимый канал, и настроить ключевые слова для него.\n"

	b.Sender.To(msg.PeerUser.AsInputPeer()).Text(msg.Ctx, welcomeText)
	return b.showMainPage(msg.Ctx, msg.PeerUser, msg.UserData, true)
}

func (h *Bot) registerCommands() {
	h.addCommand("/start", "Стартовая команда", h.startCommand)
}
