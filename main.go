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

	// Inicializar el modelo de UI con el StateManager
	model := ui.NewModel(app)

	p := tea.NewProgram(model)
	if err := p.Start(); err != nil {
		fmt.Printf("Al iniciar la aplicación: %v\n", err)
		app.Stop()
	}

	app.Stop()
}
