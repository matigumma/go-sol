package main

import (
	"fmt"
	"gosol/monitor"
	"gosol/types"
	"gosol/ui"
	"os"

	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	hello() // Call the hello function

	// Crea canales
	tokenUpdates := make(chan []types.TokenInfo)
	statusUpdates := make(chan monitor.StatusMessage)

	monitor := monitor.NewMonitor(tokenUpdates, statusUpdates)

	// go func() {
	// 	// Enviar un token de ejemplo al canal tokenUpdates
	// 	mockToken := []types.TokenInfo{
	// 		{Symbol: "MOCK", Address: "85HveQ18FegDyKnqo9evQHtUHeDt11GQYgVkse2Rpump", CreatedAt: "00:00", Score: 1000},
	// 		{Symbol: "MOCK2", Address: "85HveQ18FegDyKnqo9evQHtUHeDt11GQYgVkse2Rpump", CreatedAt: "00:02", Score: 2000},
	// 		{Symbol: "MOCK3", Address: "85HveQ18FegDyKnqo9evQHtUHeDt11GQYgVkse2Rpump", CreatedAt: "00:03", Score: 3000},
	// 		{Symbol: "MOCK4", Address: "85HveQ18FegDyKnqo9evQHtUHeDt11GQYgVkse2Rpump", CreatedAt: "00:04", Score: 4000},
	// 		{Symbol: "MOCK5", Address: "85HveQ18FegDyKnqo9evQHtUHeDt11GQYgVkse2Rpump", CreatedAt: "00:05", Score: 5000},
	// 		{Symbol: "MOCK6", Address: "85HveQ18FegDyKnqo9evQHtUHeDt11GQYgVkse2Rpump", CreatedAt: "00:06", Score: 6000},
	// 		{Symbol: "MOCK7", Address: "85HveQ18FegDyKnqo9evQHtUHeDt11GQYgVkse2Rpump", CreatedAt: "00:07", Score: 7000},
	// 		{Symbol: "MOCK8", Address: "85HveQ18FegDyKnqo9evQHtUHeDt11GQYgVkse2Rpump", CreatedAt: "00:08", Score: 8000},
	// 		{Symbol: "MOCK9", Address: "85HveQ18FegDyKnqo9evQHtUHeDt11GQYgVkse2Rpump", CreatedAt: "00:09", Score: 9000},
	// 		{Symbol: "MOCK10", Address: "85HveQ18FegDyKnqo9evQHtUHeDt11GQYgVkse2Rpump", CreatedAt: "00:10", Score: 10000},
	// 		{Symbol: "MOCK11", Address: "85HveQ18FegDyKnqo9evQHtUHeDt11GQYgVkse2Rpump", CreatedAt: "00:11", Score: 11000},
	// 		{Symbol: "MOCK12", Address: "85HveQ18FegDyKnqo9evQHtUHeDt11GQYgVkse2Rpump", CreatedAt: "00:12", Score: 12000},
	// 		{Symbol: "MOCK13", Address: "85HveQ18FegDyKnqo9evQHtUHeDt11GQYgVkse2Rpump", CreatedAt: "00:13", Score: 13000},
	// 	}
	// 	tokenUpdates <- mockToken

	// }()

	go monitor.Run()

	uiModel, _ := ui.InitProject(monitor)

	p := tea.NewProgram(uiModel, tea.WithAltScreen())
	// p := tea.NewProgram(uiModel)

	go func() {
		for tokens := range tokenUpdates {
			p.Send(ui.TokenUpdateMsg(tokens))
		}
	}()

	go func() {
		for status := range statusUpdates {
			p.Send(ui.StatusBarUpdateMsg(status))
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
