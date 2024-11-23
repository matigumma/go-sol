package ui

import (
	"gosol/monitor"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type StatusListModel struct {
	list list.Model
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func NewStatusListModel(messages []monitor.StatusMessage) StatusListModel {
	items := make([]list.Item, len(messages))
	for i, msg := range messages {
		items[i] = listItem{message: msg}
	}

	l := list.New(items, list.NewDefaultDelegate(), 100, 1) // Ajusta el tamaño según sea necesario
	l.Title = l.Styles.Spinner.Render()
	l.SetShowTitle(true)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))

	return StatusListModel{list: l}
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

func (m StatusListModel) View() string {
	return m.list.View()
}
