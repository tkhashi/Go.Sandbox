package model

import (
	"fmt"
)

// DirNode はディレクトリ構造を表現するノード
type DirNode struct {
	Name         string     // ディレクトリ/ファイル名
	AbsolutePath string     // 絶対パス
	Size         int64      // サイズ（バイト）
	RelativeSize float64    // 親ディレクトリに対する相対サイズ（0-100）
	Children     []*DirNode // 子ノード
	IsDir        bool       // ディレクトリかどうか
	IsExpanded   bool       // 展開されているかどうか
	Parent       *DirNode   // 親ノード
	IsSelected   bool       // 選択されているかどうか
}

// NewDirNode は新しいDirNodeを作成する
func NewDirNode(name, path string, isDir bool) *DirNode {
	return &DirNode{
		Name:         name,
		AbsolutePath: path,
		IsDir:        isDir,
		Children:     make([]*DirNode, 0),
		IsExpanded:   false,
	}
}

// AddChild は子ノードを追加する
func (n *DirNode) AddChild(child *DirNode) {
	child.Parent = n
	n.Children = append(n.Children, child)
}

// Toggle はノードの展開/折りたたみ状態を切り替える
func (n *DirNode) Toggle() {
	if n.IsDir {
		n.IsExpanded = !n.IsExpanded
	}
}

// CalculateRelativeSizes は親ディレクトリに対する相対サイズを計算する
func (n *DirNode) CalculateRelativeSizes() {
	if n.Parent == nil {
		n.RelativeSize = 100.0
	}

	if len(n.Children) == 0 {
		return
	}

	totalSize := n.Size
	if totalSize == 0 {
		return
	}

	for _, child := range n.Children {
		child.RelativeSize = float64(child.Size) / float64(totalSize) * 100.0
		child.CalculateRelativeSizes()
	}
}

// FormatSize はサイズを人間が読みやすい形式に変換する
func FormatSize(size int64) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
		TB
	)

	var (
		suffix  string
		divisor float64
	)

	switch {
	case size >= int64(TB):
		suffix = "TB"
		divisor = TB
	case size >= int64(GB):
		suffix = "GB"
		divisor = GB
	case size >= int64(MB):
		suffix = "MB"
		divisor = MB
	case size >= int64(KB):
		suffix = "KB"
		divisor = KB
	default:
		return "0B"
	}

	return sprintf("%.1f%s", float64(size)/divisor, suffix)
}

// sprintf はfmt.Sprintfのラッパー
func sprintf(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}
