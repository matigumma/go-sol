package main

import (
	"fmt"
	"gosol/monitor"
)

func main() {
	client, err := monitor.ConnectToWebSocket()
	if err != nil {
		fmt.Println("Error al conectar al WebSocket:", err)
		return
	}
	defer client.Close()

	pubkey := "7YttLkHDoNj9wyDur5pM1ejNaAvT9X4eqaYcHQqtj2G5"
	if err := monitor.SubscribeToLogs(client, pubkey); err != nil {
		fmt.Println("Error al suscribirse a los logs:", err)
	}
}
