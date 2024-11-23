package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"gosol/monitor"
)

type StatusListModel struct {
	list list.Model
}

func NewStatusListModel(messages []monitor.StatusMessage) StatusListModel {
	items := make([]list.Item, len(messages))
	for i, msg := range messages {
		items[i] = listItem{msg}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Status History"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

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
