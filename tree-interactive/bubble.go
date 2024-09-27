package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type errMsg error

type model struct {
	err      error
	nodes    []Noder
	quitting bool
}

var quitKeys = key.NewBinding(
	key.WithKeys("q", "esc", "ctrl+c"),
	key.WithHelp("", "press q to quit"),
)

func initialModel() model {
	dir, err := os.Getwd() // カレントディレクトリ情報取得
	if err != nil {
		return model{err: err}
	}

	nodeInfos, err := os.ReadDir(dir)
	var nodes []Noder
	for _, ni := range nodeInfos {
		switch mode := ni.Type(); {
		case mode.IsRegular():
			nodes = append(nodes, File{name: ni.Name()})
		case mode.IsDir():
			nodes = append(nodes, Directory{name: ni.Name()})
		}
	}

	return model{nodes: nodes}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			m.quitting = true
			return m, tea.Quit

		}
		return m, nil
	case errMsg:
		m.err = msg
		return m, nil

	default:
		var cmd tea.Cmd
		return m, cmd
	}
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	var str string
	for _, n := range m.nodes {
		str = str + n.GetName() + "\n"
	}

	if m.quitting {
		return str + "\n"
	}
	return str
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
