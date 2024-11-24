package main

import (
	"context"
	"os"
	"os/signal"
	"tg_reader_bot/internal/app"
	"tg_reader_bot/internal/bot"
	"tg_reader_bot/internal/client"
	"tg_reader_bot/internal/config"
	"tg_reader_bot/internal/database"
)

func main() {
	config, err := config.Init()
	if err != nil {
		panic(err)
	}

	db, err := database.Init(config.GetDatabaseQuery())
	if err != nil {
		panic(err)
	}

	app := app.GetContainer()
	app.Init(config, db)

	context, cancel := context.WithCancel(context.Background())

	go bot.Run(context)
	go client.Run(context)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	cancel()
}
