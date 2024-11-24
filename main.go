package main

import (
	"fmt"
	"gosol/monitor"
	"gosol/telegramadapter"
	"gosol/ui"

	tea "github.com/charmbracelet/bubbletea"
	// Otros imports necesarios
)

func main() {
	app := monitor.NewApp()
	fmt.Println("Iniciando la aplicación...")
	// app.Run()

	// Inicializar el modelo de UI con el StateManager
	model := ui.NewModel(app)
	fmt.Println("Modelo de UI inicializado")

	telegramclient := telegramadapter.NewTelegramClient(app)
	go telegramclient.Run()
	fmt.Println("Telegram Client inicializado")

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Al iniciar la aplicación: %v\n", err)
		app.Stop()
	}

	app.Stop()
}
