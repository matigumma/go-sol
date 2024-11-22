package main

import (
	"fmt"
	"gosol/monitor"
	"gosol/ui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// monitor.Run()

	tokens := []monitor.TokenInfo{
		{Symbol: "SOL", Address: "7YttLkHDoNj9wyDur5pM1ejNaAvT9X4eqaYcHQqtj2G5", CreatedAt: "2024-11-22", Score: 1000},
		// Add more tokens as needed
	}

	model := ui.NewModel(tokens)

	if _, err := tea.NewProgram(model).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
