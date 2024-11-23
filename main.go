package main

import (
	"fmt"
	"gosol/monitor"
	"gosol/types"
	"gosol/ui"

	tea "github.com/charmbracelet/bubbletea"
	// Otros imports necesarios
)

func main() {
	app := monitor.NewApp()
	app.Run()

	tokens := []types.TokenInfo{} // Inicialmente vacío
	ui.InitProject(app.currentMonitor)

	model := ui.NewModel(tokens, app.statusUpdates, app.tokenUpdates)

	p := tea.NewProgram(model)
	if err := p.Start(); err != nil {
		fmt.Printf("Al iniciar la aplicación: %v\n", err)
		app.Stop()
	}

	app.Stop()
}
