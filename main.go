package main

import (
	"fmt"
	"gosol/monitor"
	"gosol/types"
	"gosol/ui"
	"os"

	"log"
	"github.com/joho/godotenv"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	statusUpdates := make(chan string)

	monitor := monitor.NewMonitor(tokenUpdates, statusUpdates)
	go func() {
		// Enviar un token de ejemplo al canal tokenUpdates
		mockToken := []types.TokenInfo{
			{
				Symbol:    "MOCK",
				Address:   "85HveQ18FegDyKnqo9evQHtUHeDt11GQYgVkse2Rpump",
				CreatedAt: "00:00",
				Score:     1000,
			},
		}
		tokenUpdates <- mockToken

		monitor.Run()
	}()

	uiModel := ui.NewModel([]types.TokenInfo{})

	p := tea.NewProgram(uiModel, tea.WithAltScreen())

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
