package main

import (
	"fmt"
	"gosol/wsmanager"
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

	// Crear canales
	tokenUpdates := make(chan []types.TokenInfo)
	statusUpdates := make(chan monitor.StatusMessage)

	// Inicializar el WSManager
	wsMgr := wsmanager.NewWSManager(os.Getenv("WEBSOCKET_URL"), os.Getenv("API_KEY"), statusUpdates, tokenUpdates)
	err = wsMgr.Connect()
	if err != nil {
		log.Fatalf("WebSocket connection failed: %v", err)
	}

	// Suscribirse a los logs
	pubkey := "7YttLkHDoNj9wyDur5pM1ejNaAvT9X4eqaYcHQqtj2G5" // ray_fee_pubkey
	err = wsMgr.SubscribeToLogs(pubkey)
	if err != nil {
		log.Fatalf("Failed to subscribe to logs: %v", err)
	}

	uiModel, _ := ui.InitProject(monitor)

	p := tea.NewProgram(uiModel, tea.WithAltScreen())
	// p := tea.NewProgram(uiModel)

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
	// Cerrar el WSManager al finalizar
	wsMgr.Close()
}
