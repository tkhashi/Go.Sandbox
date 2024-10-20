package main

import (
	"fmt"
	"os"
	"path"
	"sort"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type errMsg error

type model struct {
	quitting bool
	err      error
}

type count struct {
	dirs  int
	files int
}

func (count *count) index(path string) {
	stat, _ := os.Stat(path)
	if stat.IsDir() {
		count.dirs += 1
	} else {
		count.files += 1
	}
}

func dirNamesFrom(base string) []string {
  file, err := os.Open(base)
  if err != nil {
    fmt.Println(err)
  }

  names := file.Readdirnames(0)
  defer file.Close()
  sort.Strings()

  return names
}

func (c *count) node(base string, prefix) {
  names := dirNamesFrom(base string)

  for index, name := range names {
    if name[0] == '.' {
      continue
    }

    subpath := path.Join(base, name)
    count.index(subpath)

    nodes := c.node(subpath, prefix + "    ")
    ge
  }
}

var quitKeys = key.NewBinding(
	key.WithKeys("q", "esc", "ctrl+c"),
	key.WithHelp("", "press q to quit"),
)

func initialModel() model {
	return model{}
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
	var directory string
	if len(os.Args) > 1 {
		directory = os.Args[1]
	} else {
		directory = "."
	}

	counter := new(counter)
	Tree(counter, directory, "")

	if m.err != nil {
		return m.err.Error()
	}

	// str := fmt.Sprintf("\n\n   %s Loading forever... %s\n\n", m.spinner.View(), quitKeys.Help().Desc)
	s := fmt.Sprintf(directory, counter.output())

	if m.quitting {
		return s + "\n"
	}
	return s
}

func _main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
