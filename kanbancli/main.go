package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type status int

const divisor = 4

const (
	todo status = iota
	inProgress
	done
)

// STYLING
var (
	columnStyle = lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.HiddenBorder())
	forcusedStyle = lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))
	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
)

// CUSTOM ITEM
type Task struct {
	status      status
	title       string
	description string
}

// implement the list. Item interface
func (t Task) FilterValue() string {
	return t.title
}

func (t Task) Title() string {
	return t.title
}

func (t Task) Description() string {
	return t.description
}

// MANI MODEL
type Model struct {
	forcused status
	lists []list.Model
	err  error
	loaded bool
	quitting bool
}

func New() *Model {
	return &Model{}
}

func (m *Model) Next() {
	if m.forcused == done {
		m.forcused = todo
	} else {
		m.forcused++
	}
}

func (m *Model) Prev() {
	if m.forcused == todo {
		m.forcused = done
	} else {
		m.forcused--
	}
}

// TODO: call thi on tea.WindowSizeMsg
func (m *Model) initList(width, height int) {
	defaultList := list.New([]list.Item{}, list.NewDefaultDelegate(), width/divisor, height/2)
	defaultList.SetShowHelp(false)
	m.lists = []list.Model{defaultList, defaultList, defaultList, defaultList}

	// Init To Do
	m.lists[todo].Title = "To Do"
	m.lists[todo].SetItems([]list.Item{
		Task{status: todo, title: "buy milk", description: "strawberyry milk"},
		Task{status: todo, title: "eat sushi", description: "negitoro roll, miso soup"},
		Task{status: todo, title: "fold laundry", description: "or wear wrinkly clothes"},
	})

	// Init InProgress
	m.lists[inProgress].Title = "In Progress"
	m.lists[inProgress].SetItems([]list.Item{
		Task{status: inProgress, title: "stay cool", description: "as as cucumber"},
	})

	// Init done
	m.lists[done].Title = "Done"
	m.lists[done].SetItems([]list.Item{
		Task{status: done, title: "write code", description: "don't worry wrriten in go"},
	})
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.loaded {
			columnStyle.Width(msg.Width / divisor)
			forcusedStyle.Width(msg.Width / divisor)
			columnStyle.Height(msg.Height)
			forcusedStyle.Height(msg.Height - divisor)
			m.initList(msg.Width, msg.Height)
			m.loaded = true
		}
	case tea.KeyMsg:
		switch msg.Type.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		// TODO: Why can't send keymsg without right or left
		case "left", "h":
			m.Prev()
		case "right", "l":
			m.Next()
		}
	}
	var cmd tea.Cmd
	m.lists[m.forcused], cmd = m.lists[m.forcused].Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	if m.loaded {
		todoView := m.lists[todo].View()
		inProgressView := m.lists[inProgress].View()
		doneView := m.lists[done].View()
		
		switch m.forcused {
		case inProgress:
			return lipgloss.JoinHorizontal(
				lipgloss.Left,
				columnStyle.Render(todoView),
				forcusedStyle.Render(inProgressView),
				columnStyle.Render(doneView),
			)
		case done:
			return lipgloss.JoinHorizontal(
				lipgloss.Left,
				columnStyle.Render(todoView),
				columnStyle.Render(inProgressView),
				forcusedStyle.Render(doneView),
			)
		default:
			return lipgloss.JoinHorizontal(
				lipgloss.Left,
				forcusedStyle.Render(todoView),
				columnStyle.Render(inProgressView),
				columnStyle.Render(doneView),
			)
		}
	} else {
		return "loading.."
	}
	//return m.lists[m.forcused].View()
}

func main() {
	m := New()
	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
