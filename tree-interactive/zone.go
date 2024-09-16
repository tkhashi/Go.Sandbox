package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	id    string
	title string
	desc  string
}

func (i item) Title() string       { return zone.Mark(i.id, i.title) }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return zone.Mark(i.id, i.title) }

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonWheelUp {
			m.list.CursorUp()
			return m, nil
		}

		if msg.Button == tea.MouseButtonWheelDown {
			m.list.CursorDown()
			return m, nil
		}

		if msg.Action == tea.MouseActionRelease && msg.Button == tea.MouseButtonLeft {
			for i, listItem := range m.list.VisibleItems() {
				v, _ := listItem.(item)
				if zone.Get(v.id).InBounds(msg) {
					m.list.Select(i)
					break
				}
			}
		}

		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return zone.Scan(docStyle.Render(m.list.View()))
}

func zoneMain() {
	zone.NewGlobal()

	items := []list.Item{
		item{id: "item_1", title: "Raspberry Pi’s", desc: "I have ’em all over my house"},
		item{id: "item_2", title: "Nutella", desc: "It's good on toast"},
		item{id: "item_3", title: "Bitter melon", desc: "It cools you down"},
	}

	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "Left click on an items title to select it"

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Println("error running program:", err)
		os.Exit(1)
	}
}
