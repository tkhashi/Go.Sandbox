package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/storage-visualizer/internal/model"
	"github.com/user/storage-visualizer/internal/scanner"
	"github.com/user/storage-visualizer/internal/style"
	"github.com/user/storage-visualizer/internal/ui"
)

// App はアプリケーションのメインモデル
type App struct {
	model.Model
	scanner  *scanner.Scanner
	styles   *style.Styles
	treeView *ui.TreeView
}

// NewApp は新しいAppを作成する
func NewApp() *App {
	s := style.NewStyles()
	m := model.NewModel()

	return &App{
		Model:    m,
		scanner:  scanner.NewScanner(),
		styles:   s,
		treeView: ui.NewTreeView(s),
	}
}

// Init はアプリケーションの初期化を行う
func (a *App) Init() tea.Cmd {
	// ホームディレクトリを取得
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return tea.Quit
	}

	// スキャン開始
	return a.scanDirectory(homeDir)
}

// scanDirectory はディレクトリのスキャンを開始する
func (a *App) scanDirectory(path string) tea.Cmd {
	return func() tea.Msg {
		a.Loading = true
		root, err := a.scanner.Scan(path)
		if err != nil {
			return model.ErrorMsg{Err: err}
		}
		return model.ScanCompleteMsg{Root: root}
	}
}

// Update はアプリケーションの状態を更新する
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		}

	case tea.WindowSizeMsg:
		a.Width = msg.Width
		a.Height = msg.Height
		a.treeView.SetDimensions(msg.Width, msg.Height-10, 5)

	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft {
			// クリックイベント処理
			for _, area := range a.ClickableAreas {
				if msg.X >= area.X1 && msg.X <= area.X2 && msg.Y >= area.Y1 && msg.Y <= area.Y2 {
					if area.Node != nil {
						// ノードの選択
						if a.SelectedNode != nil {
							a.SelectedNode.IsSelected = false
						}
						area.Node.IsSelected = true
						a.SelectedNode = area.Node

						// ディレクトリの場合は展開/折りたたみ
						if area.Node.IsDir {
							area.Node.Toggle()
						}

						// 現在のノードを更新
						a.CurrentNode = area.Node
						break
					}
				}
			}
		}

	case model.ScanCompleteMsg:
		a.Root = msg.Root
		a.CurrentNode = a.Root
		a.Loading = false

	case model.ErrorMsg:
		a.Error = msg.Err
		a.Loading = false
	}

	// 親モデルの更新
	var cmd tea.Cmd
	parentModel, parentCmd := a.Model.Update(msg)
	if m, ok := parentModel.(model.Model); ok {
		a.Model = m
	}
	cmd = parentCmd

	return a, cmd
}

// View はアプリケーションの表示を行う
func (a *App) View() string {
	if a.Loading {
		return a.styles.App.Render("Loading... Please wait while scanning directory structure.")
	}

	if a.Error != nil {
		return a.styles.App.Render(fmt.Sprintf("Error: %s", a.Error.Error()))
	}

	if a.Root == nil {
		return a.styles.App.Render("No data loaded")
	}

	// ヘッダー
	header := a.styles.Header.Render("Storage Visualizer - Press q to quit")

	// ツリービュー
	treeContent, clickAreas := a.treeView.Render(a.Root, a.CurrentNode)
	a.ClickableAreas = clickAreas

	// 詳細ビュー
	var detailContent string
	if a.SelectedNode != nil {
		detailContent = a.treeView.RenderProgressBar(a.SelectedNode)
	} else {
		detailContent = a.treeView.RenderProgressBar(a.Root)
	}

	// レイアウト
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		a.styles.BorderBox.Render(treeContent),
		a.styles.BorderBox.Render(detailContent),
	)

	// フッター
	footer := a.styles.Footer.Render("🖱️ Click: Select/Expand   📊 Size: " + model.FormatSize(a.Root.Size))

	// 全体のレイアウト
	return a.styles.App.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			mainContent,
			footer,
		),
	)
}

func main() {
	app := NewApp()
	p := tea.NewProgram(app, tea.WithMouseCellMotion(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running app: %v", err)
		os.Exit(1)
	}
}
