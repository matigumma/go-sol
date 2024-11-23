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

	// Crea canales
	tokenUpdates := make(chan []types.TokenInfo)
	statusUpdates := make(chan monitor.StatusMessage)

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
