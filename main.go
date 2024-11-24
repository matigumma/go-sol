package main

import (
	"fmt"
	"gosol/monitor"
	"gosol/telegramadapter"
	"gosol/ui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

var websocketURL string
var apiKey string
var pubkey string
var apiBaseURL string

func main() {
	pubkey = os.Getenv("RAY_FEE_PUBKEY")
	apiBaseURL = os.Getenv("API_BASE_URL")

	websocketURL = os.Getenv("WEBSOCKET_URL")
	apiKey = os.Getenv("API_KEY")

	app := monitor.NewApp()
	// app.Run()

	telegramclient := telegramadapter.NewTelegramClient(app)
	telegramclient.Run()

	// Inicializar el modelo de UI con el StateManager
	model := ui.NewModel(app)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Al iniciar la aplicación: %v\n", err)
		app.Stop()
	}
	// Cerrar el WSManager al finalizar
	wsMgr.Close()
}
