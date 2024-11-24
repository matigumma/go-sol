package bot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tg_reader_bot/internal/app"
	"tg_reader_bot/internal/cache"
	"tg_reader_bot/internal/events"
	"tg_reader_bot/internal/protobufs"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lj "gopkg.in/natefinch/lumberjack.v2"
)

type btnCallback func(buttonContext) error

type buttonContext struct {
	Ctx      context.Context
	Entities tg.Entities
	Update   *tg.UpdateBotCallbackQuery
	User     *tg.User
	UserData *cache.UserData
	Data     []byte
}

type commandInfo struct {
	Description string
	callback    func(events.MsgContext) error
}

type Bot struct {
	Client        *telegram.Client
	Sender        *message.Sender
	startTime     uint64
	cmdsCallbacks map[string]commandInfo
	btnCallbacks  map[protobufs.MessageID]btnCallback
	peersCache    cache.PeersManager
}

func Init(client *telegram.Client) *Bot {
	bot := &Bot{
		Client:        client,
		Sender:        message.NewSender(client.API()),
		startTime:     uint64(time.Now().Unix()),
		cmdsCallbacks: make(map[string]commandInfo),
		btnCallbacks:  make(map[protobufs.MessageID]btnCallback),
		peersCache:    cache.PeersManager{Peers: make(map[int64]*cache.PeerData), Users: make(map[int64]*cache.UserData)},
	}

	bot.registerCommands()
	bot.registerQueryCallbacks()

	return bot
}

func (b *Bot) API() *tg.Client {
	return b.Client.API()
}

func (bot *Bot) registerCommandsInBot(ctx context.Context) error {
	botCommands := bot.GetCommands()

	commands := make([]tg.BotCommand, 0, len(botCommands))
	for key, value := range botCommands {
		commands = append(commands, tg.BotCommand{
			Command:     strings.TrimPrefix(key, "/"),
			Description: value.Description,
		})
	}

	if _, err := bot.API().BotsSetBotCommands(ctx, &tg.BotsSetBotCommandsRequest{
		Scope:    &tg.BotCommandScopeDefault{},
		Commands: commands,
		LangCode: "ru",
	}); err != nil {
		return errors.Wrap(err, "register commands")
	}

	if _, err := bot.API().BotsSetBotCommands(ctx, &tg.BotsSetBotCommandsRequest{
		Scope:    &tg.BotCommandScopeDefault{},
		Commands: commands,
		LangCode: "en",
	}); err != nil {
		return errors.Wrap(err, "register commands")
	}

	return nil
}

func botRun(ctx context.Context) error {
	app := app.GetContainer()
	config := app.Config

	botDir := "bot"
	if err := os.MkdirAll(botDir, 0700); err != nil {
		return err
	}

	logFilePath := filepath.Join(botDir, "log.jsonl")
	logWriter := zapcore.AddSync(&lj.Logger{
		Filename:   logFilePath,
		MaxBackups: 3,
		MaxSize:    1, // megabytes
		MaxAge:     7, // days
	})

	logCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		logWriter,
		zap.InfoLevel,
	)
	lg := zap.New(logCore)
	defer func() { _ = lg.Sync() }()

	dispatcher := tg.NewUpdateDispatcher()

	updatesRecovery := updates.New(updates.Config{
		Handler: dispatcher,
		Logger:  lg.Named("bot.updates.recovery"),
	})

	waiter := floodwait.NewWaiter().WithCallback(func(ctx context.Context, wait floodwait.FloodWait) {
		fmt.Println("Bot API FLOOD_WAIT. Will retry after", wait.Duration)
	})

	options := telegram.Options{
		Logger:        lg,
		UpdateHandler: updatesRecovery,
		SessionStorage: &telegram.FileSessionStorage{
			Path: filepath.Join(botDir, "session.json"),
		},
		Middlewares: []telegram.Middleware{
			waiter,
		},
	}

	client := telegram.NewClient(config.AppID, config.AppHash, options)

	bot := Init(client)
	bot.UpdateHandles(dispatcher)

	api := bot.API()

	return waiter.Run(ctx, func(ctx context.Context) error {
		if err := client.Run(ctx, func(ctx context.Context) error {
			status, err := client.Auth().Status(ctx)
			if err != nil {
				return err
			}

			/* Can be already authenticated if we have valid session in session storage. */
			if !status.Authorized {
				if _, err := client.Auth().Bot(ctx, config.APIToken); err != nil {
					return errors.Wrap(err, "auth")
				}
			}

			user, err := client.Self(ctx)
			if err != nil {
				return errors.Wrap(err, "call self")
			}

			err = bot.registerCommandsInBot(ctx)
			if err != nil {
				return err
			}

			return updatesRecovery.Run(ctx, api, user.ID, updates.AuthOptions{
				IsBot: user.Bot,
				OnStart: func(ctx context.Context) {
					fmt.Println("Bot Started.")
					bot.peersCache.LoadUsersData()
					go bot.ParseChannels(ctx)
				},
			})
		}); err != nil {
			return errors.Wrap(err, "run")
		}
		return nil
	})
}

func Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Bot stopped.")
			return
		default:
			botRun(ctx)
		}
	}
}
