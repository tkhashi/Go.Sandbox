package model

import (
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

// メッセージ型
type ErrorMsg struct {
	Err error
}

type ScanCompleteMsg struct {
	Root *DirNode
}

// SortType はソートの種類を表す
type SortType int

const (
	SortBySize SortType = iota
	SortByName
)

// Model はアプリケーションの状態を表す
type Model struct {
	// データ
	Root         *DirNode
	CurrentNode  *DirNode
	SelectedNode *DirNode

	// UI状態
	Width    int
	Height   int
	SortBy   SortType
	Loading  bool
	Error    error
	Progress progress.Model

	// クリック可能領域
	ClickableAreas []ClickableArea
}

// ClickableArea はクリック可能な領域を表す
type ClickableArea struct {
	X1, Y1 int // 開始座標
	X2, Y2 int // 終了座標
	Node   *DirNode
}

// NewModel は新しいModelを作成する
func NewModel() Model {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	return Model{
		Root:           nil,
		CurrentNode:    nil,
		SelectedNode:   nil,
		SortBy:         SortBySize,
		Loading:        false,
		Progress:       p,
		ClickableAreas: make([]ClickableArea, 0),
	}
}

// Init はモデルの初期化を行う
func (m Model) Init() tea.Cmd {
	return nil
}

// Update はモデルの状態を更新する
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Progress.Width = msg.Width - 20

	case tea.MouseMsg:
		// マウスイベント処理（後で実装）
	}

	progressModel, cmd := m.Progress.Update(msg)
	m.Progress = progressModel.(progress.Model)
	return m, cmd
}

// View はモデルの表示を行う
func (m Model) View() string {
	if m.Loading {
		return "Loading..."
	}

	if m.Error != nil {
		return "Error: " + m.Error.Error()
	}

	if m.Root == nil {
		return "No data loaded"
	}

	// 実際の表示は後で実装
	return "Storage Visualizer"
}
