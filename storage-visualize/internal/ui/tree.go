package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/storage-visualizer/internal/model"
	"github.com/user/storage-visualizer/internal/style"
)

// TreeView はディレクトリツリーを表示するコンポーネント
type TreeView struct {
	styles     *style.Styles
	clickAreas []model.ClickableArea
	width      int
	height     int
	startY     int
}

// NewTreeView は新しいTreeViewを作成する
func NewTreeView(styles *style.Styles) *TreeView {
	return &TreeView{
		styles:     styles,
		clickAreas: make([]model.ClickableArea, 0),
	}
}

// SetDimensions はTreeViewの寸法を設定する
func (t *TreeView) SetDimensions(width, height, startY int) {
	t.width = width
	t.height = height
	t.startY = startY
}

// Render はツリービューを描画する
func (t *TreeView) Render(root *model.DirNode, currentNode *model.DirNode) (string, []model.ClickableArea) {
	if root == nil {
		return "No data", nil
	}

	t.clickAreas = make([]model.ClickableArea, 0)
	result := t.styles.Title.Render("📊 Storage Visualizer")
	result += "\n\n"

	// ルートディレクトリの情報
	rootInfo := fmt.Sprintf("📂 %s (%s)", root.Name, model.FormatSize(root.Size))
	result += t.styles.Directory.Render(rootInfo)
	result += "\n"

	// 子ノードを再帰的に描画
	lines := t.renderNode(root, "", true, 2)
	result += strings.Join(lines, "\n")

	return result, t.clickAreas
}

// renderNode はノードを再帰的に描画する
func (t *TreeView) renderNode(node *model.DirNode, prefix string, isLast bool, y int) []string {
	lines := []string{}

	// このノードのプレフィックス
	nodePrefix := prefix
	if isLast {
		nodePrefix += "└── "
	} else {
		nodePrefix += "├── "
	}

	// 子ノードのプレフィックス
	childPrefix := prefix
	if isLast {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}

	// 子ノードを描画
	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1

		// ノード情報
		var info string
		if child.IsDir {
			// ディレクトリの場合
			expandIcon := "📁"
			if child.IsExpanded {
				expandIcon = "📂"
			}
			info = fmt.Sprintf("%s %s (%s) [%.1f%%]",
				expandIcon,
				child.Name,
				model.FormatSize(child.Size),
				child.RelativeSize)

			// スタイル適用
			var renderedInfo string
			if child.IsSelected {
				// 選択中のノードは強調表示
				renderedInfo = t.styles.Directory.
					Copy().
					Background(lipgloss.Color(style.ColorHighlight)).
					Foreground(lipgloss.Color("#000000")).
					Render(info)
			} else {
				renderedInfo = t.styles.Directory.Render(info)
			}

			// クリック可能領域を登録
			lineY := y + len(lines)
			if lineY >= t.startY && lineY < t.startY+t.height {
				t.clickAreas = append(t.clickAreas, model.ClickableArea{
					X1:   0,
					Y1:   lineY,
					X2:   len(nodePrefix) + len(info),
					Y2:   lineY,
					Node: child,
				})
			}

			// 行を追加
			lines = append(lines, nodePrefix+renderedInfo)

			// 展開されている場合は子ノードも描画
			if child.IsExpanded {
				childLines := t.renderNode(child, childPrefix, isLastChild, y+len(lines))
				lines = append(lines, childLines...)
			}
		} else {
			// ファイルの場合
			info = fmt.Sprintf("📄 %s (%s) [%.1f%%]",
				child.Name,
				model.FormatSize(child.Size),
				child.RelativeSize)

			// スタイル適用
			var renderedInfo string
			if child.IsSelected {
				// 選択中のノードは強調表示
				renderedInfo = t.styles.File.
					Copy().
					Background(lipgloss.Color(style.ColorHighlight)).
					Foreground(lipgloss.Color("#000000")).
					Render(info)
			} else {
				renderedInfo = t.styles.File.Render(info)
			}

			// クリック可能領域を登録
			lineY := y + len(lines)
			if lineY >= t.startY && lineY < t.startY+t.height {
				t.clickAreas = append(t.clickAreas, model.ClickableArea{
					X1:   0,
					Y1:   lineY,
					X2:   len(nodePrefix) + len(info),
					Y2:   lineY,
					Node: child,
				})
			}

			// 行を追加
			lines = append(lines, nodePrefix+renderedInfo)
		}
	}

	return lines
}

// RenderProgressBar はプログレスバーを描画する
func (t *TreeView) RenderProgressBar(node *model.DirNode) string {
	if node == nil {
		return ""
	}

	result := t.styles.Title.Render("📊 Size Distribution")
	result += "\n\n"

	// 子ノードがない場合
	if len(node.Children) == 0 {
		return result + "No children"
	}

	// 子ノードのサイズ分布を表示
	for _, child := range node.Children {
		name := child.Name
		if len(name) > 15 {
			name = name[:12] + "..."
		}

		// 名前とサイズ
		info := fmt.Sprintf("%-15s %s", name, model.FormatSize(child.Size))
		result += t.styles.Directory.Render(info)
		result += "\n"

		// プログレスバー
		bar := t.styles.RenderProgressBar(child.RelativeSize, 30)
		percent := fmt.Sprintf(" %.1f%%", child.RelativeSize)
		result += bar + t.styles.Percentage.Render(percent)
		result += "\n\n"
	}

	return result
}
