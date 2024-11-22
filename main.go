package main

import (
	"fmt"
	"gosol/ui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// monitor.Run()

	content := `# Hello World

	This is a simple example of Markdown rendering with Glamour!
	Check out the [other examples](https://github.com/charmbracelet/glamour/tree/master/examples) too.

	Bye!
	`

	// r, _ := glamour.NewTermRenderer(
	// 	// detect background color and pick either the default dark or light theme
	// 	glamour.WithStylePath("dark"),
	// 	// wrap output at specific width (default is 80)
	// 	glamour.WithWordWrap(40),
	// )

	// out, err := r.Render(in)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Print(out)

	model, err := ui.NewExample(content)
	if err != nil {
		fmt.Println("Could not initialize Bubble Tea model:", err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(model).Run(); err != nil {
		fmt.Println("Bummer, there's been an error:", err)
		os.Exit(1)
	}
}
