package main

import (
	"fmt"
	"gosol/monitor"
	"gosol/ui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	tokenUpdates := make(chan []monitor.TokenInfo)

	go monitor.Run(tokenUpdates)

	model := ui.NewModel([]monitor.TokenInfo{})

	p := tea.NewProgram(model)

	go func() {
		for tokens := range tokenUpdates {
			p.Send(tokens)
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
