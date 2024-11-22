package main

import (
	"flag"
	"fmt"
	"os"

	"gosol/monitor"
	"gosol/dashboard"
)

func main() {
	mode := flag.String("mode", "monitor", "Mode to run the application: 'monitor' or 'dashboard'")
	flag.Parse()

	switch *mode {
	case "monitor":
		monitor.Run()
	case "dashboard":
		dashboard.RunDashboard(monitor.GetMintState())
	default:
		fmt.Println("Invalid mode. Use 'monitor' or 'dashboard'.")
		os.Exit(1)
	}
}
