package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea" // Bubble Tea を 'tea' としてエイリアス
)

// 定数の定義
const (
	width      = 40
	height     = 20
	tickPerSec = 5
)

// CellGrid はセルの状態を保持する2Dスライス
type CellGrid [][]bool

// Model はBubble Teaのモデル
type Model struct {
	grid     CellGrid
	running  bool
	quitting bool
	cursorX  int
	cursorY  int
}

// NewModel は初期モデルを作成
func NewModel() Model {
	grid := make(CellGrid, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			grid[y][x] = rand.Float32() < 0.2 // 20%の確率で生きているセル
		}
	}
	return Model{
		grid:    grid,
		running: false,
		cursorX: 0,
		cursorY: 0,
	}
}

// Init は初期コマンドを返す
func (m Model) Init() tea.Cmd {
	return nil
}

// Update はメッセージに応じてモデルを更新
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case " ":
			m.running = !m.running
			if m.running {
				return m, m.tick()
			}
			return m, nil

		case "up":
			if m.cursorY > 0 {
				m.cursorY--
			}
			return m, nil

		case "down":
			if m.cursorY < height-1 {
				m.cursorY++
			}
			return m, nil

		case "left":
			if m.cursorX > 0 {
				m.cursorX--
			}
			return m, nil

		case "right":
			if m.cursorX < width-1 {
				m.cursorX++
			}
			return m, nil

		case "enter":
			m.grid[m.cursorY][m.cursorX] = !m.grid[m.cursorY][m.cursorX]
			return m, nil
		}

	case time.Time:
		if m.running {
			m.grid = step(m.grid)
			return m, m.tick()
		}
	}

	return m, nil
}

// tick は次のティックをスケジュール
func (m Model) tick() tea.Cmd {
	return tea.Tick(time.Second/tickPerSec, func(t time.Time) tea.Msg {
		return t
	})
}

// View は現在のモデルを描画
func (m Model) View() string {
	if m.quitting {
		return "終了しました。\n"
	}

	s := "ライフゲーム - スペースで開始/停止, 矢印キーでカーソル移動, Enterでセルの切り替え, qで終了\n\n"

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x == m.cursorX && y == m.cursorY {
				s += "["
			} else {
				s += " "
			}

			if m.grid[y][x] {
				s += "O"
			} else {
				s += " "
			}

			if x == m.cursorX && y == m.cursorY {
				s += "]"
			} else {
				s += " "
			}
		}
		s += "\n"
	}

	status := "状態: "
	if m.running {
		status += "実行中"
	} else {
		status += "停止中"
	}
	s += "\n" + status

	return s
}

// step は次の世代のグリッドを計算
func step(current CellGrid) CellGrid {
	next := make(CellGrid, height)
	for y := 0; y < height; y++ {
		next[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			live := countLiveNeighbors(current, x, y)
			if current[y][x] {
				next[y][x] = live == 2 || live == 3
			} else {
				next[y][x] = live == 3
			}
		}
	}
	return next
}

// countLiveNeighbors は指定したセルの周囲の生きているセルの数をカウント
func countLiveNeighbors(grid CellGrid, x, y int) int {
	count := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height && grid[ny][nx] {
				count++
			}
		}
	}
	return count
}

func main() {
	rand.Seed(time.Now().UnixNano())
	p := tea.NewProgram(NewModel())

	if err := p.Start(); err != nil {
		fmt.Printf("プログラム終了エラー: %v\n", err)
		os.Exit(1)
	}
}
