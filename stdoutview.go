package ui

import (
	"bufio"
	"io"
	"os/exec"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type stdoutMsg string
type stdoutErrMsg struct{ err error }
type stdoutDoneMsg struct{}

type StdoutViewModel struct {
	sub      chan string
	viewport viewport.Model
}

func NewStdoutViewModel() StdoutViewModel {
	return StdoutViewModel{
		sub:      make(chan string),
		viewport: viewport.New(0, viewportHeight),
	}
}

func (m *StdoutViewModel) Init() tea.Cmd {
	return tea.Batch(executeCommand(m.sub), waitForStdoutResponses(m.sub))
}

func executeCommand(sub chan string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("your-command", "arg1", "arg2") // Reemplaza con tu comando
		out, err := cmd.StdoutPipe()
		if err != nil {
			return stdoutErrMsg{err}
		}

		if err := cmd.Start(); err != nil {
			return stdoutErrMsg{err}
		}

		buf := bufio.NewReader(out)
		for {
			line, _, err := buf.ReadLine()
			if err == io.EOF {
				return stdoutDoneMsg{}
			}
			if err != nil {
				return stdoutErrMsg{err}
			}
			sub <- string(line)
		}
	}
}

func waitForStdoutResponses(sub chan string) tea.Cmd {
	return func() tea.Msg {
		return stdoutMsg(<-sub)
	}
}

func (m *StdoutViewModel) Update(msg tea.Msg) (tea.Cmd, bool) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case stdoutMsg:
		m.appendOutput(string(msg))
		cmd = waitForStdoutResponses(m.sub)
	case stdoutErrMsg:
		m.appendOutput("Error: " + msg.err.Error())
	case stdoutDoneMsg:
		m.appendOutput("Done.")
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return cmd, true
}

func (m *StdoutViewModel) appendOutput(s string) {
	m.viewport.SetContent(m.viewport.Content() + "\n" + s)
	m.viewport.GotoBottom()
}

func (m StdoutViewModel) View() string {
	return m.viewport.View()
}
