package ui

import (
	"gosol/monitor"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type StatusListModel struct {
	list    list.Model
	spinner spinner.Model
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func NewStatusListModel(messages []monitor.StatusMessage) StatusListModel {
	items := make([]list.Item, len(messages))
	for i, msg := range messages {
		items[len(messages)-1-i] = listItem{message: msg} // Invertir el orden de los mensajes
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	l := list.New(items, list.NewDefaultDelegate(), 180, 1) // Ajusta el tamaño según sea necesario
	l.Title = s.View()
	l.SetShowTitle(true)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))

	return StatusListModel{list: l, spinner: s}
}

type listItem struct {
	message monitor.StatusMessage
}

func (i listItem) Title() string {
	return i.message.Message
}

func (i listItem) Description() string {
	return ""
}

func (i listItem) FilterValue() string {
	return i.message.Message
}

func (m *StatusListModel) Update(msg tea.Msg) (tea.Cmd, bool) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return cmd, true
}

func (m StatusListModel) View() string {
	m.list.Title = m.spinner.View() // Actualizar el título con la vista del spinner
	return m.list.View()
}
