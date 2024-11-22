package main

import (
	"bufio"
	"fmt"
	"os"

	"gosol/monitor"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Select mode to run the application: 'monitor' or 'dashboard'")
	fmt.Print("Enter mode: ")
	mode, _ := reader.ReadString('\n')
	mode = mode[:len(mode)-1] // Remove newline character

	switch *mode {
	case "monitor":
		monitor.Run()
	case "dashboard":
		monitor.RunDashboard(monitor.GetMintState())
	default:
		fmt.Println("Invalid mode. Use 'monitor' or 'dashboard'.")
		os.Exit(1)
	}
}
