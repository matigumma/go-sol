package main

import (
	"fmt"
	"gosol/monitor"
	"gosol/types"
	"gosol/ui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	tokenUpdates := make(chan []types.TokenInfo)

	statusUpdates := make(chan string)

	monitor := monitor.NewMonitor(tokenUpdates, statusUpdates)
	go func() {
		// Enviar un token de ejemplo al canal tokenUpdates
		mockToken := []types.TokenInfo{
			{
				Symbol:    "MOCK",
				Address:   "MockAddress",
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
			fmt.Println("Received token updates:", tokens) // Log statement added
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
