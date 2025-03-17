package style

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles はアプリケーション全体のスタイルを定義する
type Styles struct {
	// ベーススタイル
	App    lipgloss.Style
	Header lipgloss.Style
	Footer lipgloss.Style

	// ボックススタイル
	BorderBox  lipgloss.Style
	ContentBox lipgloss.Style

	// テキストスタイル
	Title      lipgloss.Style
	Directory  lipgloss.Style
	File       lipgloss.Style
	Size       lipgloss.Style
	Percentage lipgloss.Style

	// プログレスバースタイル
	ProgressFull  lipgloss.Style
	ProgressEmpty lipgloss.Style
}

// カラーパレット
const (
	ColorPrimary   = "#5D4FE3" // メインカラー
	ColorSecondary = "#2EC4B6" // アクセントカラー
	ColorBorder    = "#4A4A4A" // ボーダー
	ColorText      = "#FFFFFF" // テキスト
	ColorHighlight = "#FFD23F" // ハイライト
	ColorError     = "#FF5252" // エラー
)

// NewStyles は新しいStylesを作成する
func NewStyles() *Styles {
	s := new(Styles)

	// ベーススタイル
	s.App = lipgloss.NewStyle().
		Padding(1, 2)

	s.Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorPrimary)).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Padding(0, 1)

	s.Footer = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText)).
		Background(lipgloss.Color(ColorPrimary)).
		Padding(0, 1)

	// ボックススタイル
	s.BorderBox = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Padding(1, 2)

	s.ContentBox = lipgloss.NewStyle().
		Padding(1, 2)

	// テキストスタイル
	s.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorPrimary)).
		MarginBottom(1)

	s.Directory = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSecondary)).
		Bold(true)

	s.File = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText))

	s.Size = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorHighlight))

	s.Percentage = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorHighlight)).
		Bold(true)

	// プログレスバースタイル
	s.ProgressFull = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Background(lipgloss.Color(ColorPrimary))

	s.ProgressEmpty = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorBorder)).
		Background(lipgloss.Color(ColorBorder))

	return s
}

// RenderProgressBar はプログレスバーを描画する
func (s *Styles) RenderProgressBar(percent float64, width int) string {
	// パーセンテージを0-100の範囲に制限
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}

	// プログレスバーの幅を計算
	filled := int(float64(width) * percent / 100)
	empty := width - filled

	// プログレスバーを描画
	bar := ""
	if filled > 0 {
		bar += s.ProgressFull.Render(repeat("█", filled))
	}
	if empty > 0 {
		bar += s.ProgressEmpty.Render(repeat("░", empty))
	}

	return bar
}

// repeat は文字列を指定回数繰り返す
func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
