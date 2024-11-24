package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"tg_reader_bot/internal/app"
	"tg_reader_bot/internal/events"

	tgt "tg_reader_bot/internal/telegram"

	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lj "gopkg.in/natefinch/lumberjack.v2"
)

func clientRun(ctx context.Context) error {
	app := app.GetContainer()
	config := app.Config

	clientDir := "client"
	if err := os.MkdirAll(clientDir, 0700); err != nil {
		return err
	}

	logFilePath := filepath.Join(clientDir, "log.jsonl")
	logWriter := zapcore.AddSync(&lj.Logger{
		Filename:   logFilePath,
		MaxBackups: 3,
		MaxSize:    1,
		MaxAge:     7,
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
		Logger:  lg.Named("client.updates.recovery"),
		Handler: dispatcher,
	})

	waiter := floodwait.NewWaiter().WithCallback(func(ctx context.Context, wait floodwait.FloodWait) {
		fmt.Println("Client API FLOOD_WAIT. Will retry after", wait.Duration)
	})

	options := telegram.Options{
		Logger:        lg,
		UpdateHandler: updatesRecovery,
		SessionStorage: &telegram.FileSessionStorage{
			Path: filepath.Join(clientDir, "session.json"),
		},
		Middlewares: []telegram.Middleware{
			waiter,
		},
	}

	client := telegram.NewClient(config.AppID, config.AppHash, options)
	app.Client = tgt.InitTGClient(client)

	flow := auth.NewFlow(
		events.Auth{PhoneNumber: config.PhoneNumber},
		auth.SendCodeOptions{},
	)

	return waiter.Run(ctx, func(ctx context.Context) error {
		if err := client.Run(ctx, func(ctx context.Context) error {
			if err := client.Auth().IfNecessary(ctx, flow); err != nil {
				return err
			}

			user, err := client.Self(ctx)
			if err != nil {
				return errors.Wrap(err, "call self")
			}

			return updatesRecovery.Run(ctx, client.API(), user.ID, updates.AuthOptions{
				IsBot: user.Bot,
				OnStart: func(ctx context.Context) {
					fmt.Println("Client Started.")
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
			fmt.Println("Client stopped.")
			return
		default:
			clientRun(ctx)
		}
	}
}
