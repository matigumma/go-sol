package main

import (
	"fmt"
	"gosol/monitor"
	"gosol/telegramadapter"
	"gosol/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	app := monitor.NewApp()
	app.Run()

	// Inicializar el modelo de UI con el StateManager
	model := ui.NewModel(app)

	telegramclient := telegramadapter.NewTelegramClient(app)
	go telegramclient.Run()

	// p := tea.NewProgram(model, tea.WithAltScreen())
	p := tea.NewProgram(model)
	// p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Al iniciar la aplicación: %v\n", err)
		app.Stop()
	}

	app.Stop()

}
