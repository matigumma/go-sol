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

	monitor := monitor.NewMonitor(tokenUpdates)
	go monitor.Run()

	uiModel := ui.NewModel([]types.TokenInfo{})

	p := tea.NewProgram(uiModel, tea.WithAltScreen())

	go func() {
		for tokens := range tokenUpdates {
			fmt.Println("Received token updates:", tokens) // Log statement added
			p.Send(ui.TokenUpdateMsg(tokens))
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
