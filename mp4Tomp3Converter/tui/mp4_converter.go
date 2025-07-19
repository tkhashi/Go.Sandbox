package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// スタイル定義
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Margin(1, 0)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))

	subtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Margin(1, 0)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			Margin(0, 0, 1, 0)
)

// アプリケーションの状態
type AppState int

const (
	StateSelectingFolder AppState = iota
	StateConfirmSelection
	StateConverting
	StateDone
)

// ConversionJob 変換ジョブ
type ConversionJob struct {
	InputPath  string
	OutputPath string
	Status     string
	Error      error
}

// Model 全体のアプリケーションモデル
type Model struct {
	state          AppState
	filepicker     filepicker.Model
	selectedPath   string
	jobs           []ConversionJob
	current        int
	progress       progress.Model
	spinner        spinner.Model
	err            error
	mp4Count       int
	completedCount int // 完了したジョブ数
	maxConcurrent  int // 最大並列数
	initialDir     string // 実行時のディレクトリ
}

// メッセージ定義
type JobDoneMsg struct {
	index int
	err   error
}

type AllDoneMsg struct{}

type FolderSelectedMsg struct {
	path     string
	mp4Count int
}

type ConversionStartMsg struct {
	jobs []ConversionJob
}

type DirectoryChangedMsg struct {
	path string
}

func initialModel() Model {
	// 実行時のディレクトリを保存
	initialDir, _ := os.Getwd()
	
	// ファイルピッカーの設定
	fp := filepicker.New()
	fp.AllowedTypes = []string{} // すべてのファイル・フォルダを表示
	fp.CurrentDirectory = initialDir
	fp.ShowHidden = false
	fp.DirAllowed = true
	fp.FileAllowed = true // ファイルも表示する

	// プログレスバーの設定
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	// スピナーの設定
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		state:         StateSelectingFolder,
		filepicker:    fp,
		progress:      p,
		spinner:       s,
		maxConcurrent: 3, // 3つまで並列実行
		initialDir:    initialDir,
	}
}

func (m Model) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateSelectingFolder:
		return m.updateFolderSelection(msg)
	case StateConfirmSelection:
		return m.updateConfirmation(msg)
	case StateConverting:
		return m.updateConversion(msg)
	case StateDone:
		return m.updateDone(msg)
	}
	return m, nil
}

func (m Model) updateFolderSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "h", "left":
			// 親ディレクトリに移動
			return m, m.navigateToParent()
		case "l", "right":
			// 選択されたフォルダに入る
			selectedPath := m.filepicker.Path
			if selectedPath != "" {
				if info, err := os.Stat(selectedPath); err == nil && info.IsDir() {
					// ディレクトリの場合は入る
					return m, m.navigateToDirectory(selectedPath)
				}
			}
		case "enter":
			// 選択されたフォルダに入るか、現在のディレクトリを選択
			selectedPath := m.filepicker.Path
			if selectedPath != "" {
				if info, err := os.Stat(selectedPath); err == nil && info.IsDir() {
					// ディレクトリの場合は入る
					return m, m.navigateToDirectory(selectedPath)
				}
			} else {
				// 何も選択されていない場合は現在のディレクトリを選択
				return m, m.checkMP4Files(m.filepicker.CurrentDirectory)
			}
		case "space":
			// 現在のディレクトリを選択
			if m.filepicker.CurrentDirectory != "" {
				return m, m.checkMP4Files(m.filepicker.CurrentDirectory)
			}
		case "p":
			// 親ディレクトリに移動（追加ショートカット）
			return m, m.navigateToParent()
		case "~":
			// ホームディレクトリに移動
			if homeDir, err := os.UserHomeDir(); err == nil {
				return m, m.navigateToDirectory(homeDir)
			}
		case "/":
			// ルートディレクトリに移動
			return m, m.navigateToDirectory("/")
		case ".":
			// 実行時のディレクトリに移動
			return m, m.navigateToDirectory(m.initialDir)
		}

	case FolderSelectedMsg:
		m.selectedPath = msg.path
		m.mp4Count = msg.mp4Count
		m.state = StateConfirmSelection
		return m, nil

	case DirectoryChangedMsg:
		// ファイルピッカーのディレクトリを変更
		m.filepicker.CurrentDirectory = msg.path
		// ファイルピッカーを再読み込み
		return m, tea.Batch(
			m.filepicker.Init(),
			func() tea.Msg { return nil }, // ダミーメッセージでView更新をトリガー
		)
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	// ディレクトリが選択された場合
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		if info, err := os.Stat(path); err == nil {
			if info.IsDir() {
				// ディレクトリの場合は入る
				return m, m.navigateToDirectory(path)
			} else {
				// ファイルの場合は何もしない（または親ディレクトリを選択）
				return m, m.checkMP4Files(m.filepicker.CurrentDirectory)
			}
		}
	}

	return m, cmd
}

func (m Model) navigateToParent() tea.Cmd {
	return func() tea.Msg {
		parentDir := filepath.Dir(m.filepicker.CurrentDirectory)
		if parentDir != m.filepicker.CurrentDirectory { // 既にルートでない場合
			return DirectoryChangedMsg{path: parentDir}
		}
		return nil
	}
}

func (m Model) navigateToDirectory(path string) tea.Cmd {
	return func() tea.Msg {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return DirectoryChangedMsg{path: path}
		}
		return nil
	}
}

func (m Model) checkMP4Files(path string) tea.Cmd {
	return func() tea.Msg {
		files, _ := findMP4Files(path)
		return FolderSelectedMsg{
			path:     path,
			mp4Count: len(files),
		}
	}
}

func (m Model) updateConfirmation(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "y", "Y", "enter":
			m.state = StateConverting  // 状態遷移を先に設定
			return m, m.startConversion()
		case "n", "N", "esc":
			m.state = StateSelectingFolder
			return m, nil
		}
	case ConversionStartMsg:
		// 変換開始メッセージも処理
		m.jobs = msg.jobs
		m.current = 0
		return m, tea.Batch(
			m.spinner.Tick,
			m.convertNext(),
		)
	}
	return m, nil
}

func (m Model) startConversion() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// FFmpegの確認
		if !checkFFmpeg() {
			return fmt.Errorf("ffmpeg not found")
		}

		// MP4ファイルの検索
		mp4Files, err := findMP4Files(m.selectedPath)
		if err != nil {
			return err
		}

		if len(mp4Files) == 0 {
			return fmt.Errorf("no MP4 files found")
		}

		// outputディレクトリの作成
		outputDir := filepath.Join(m.selectedPath, "output")
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return err
		}

		// ジョブの作成
		var jobs []ConversionJob
		for _, mp4File := range mp4Files {
			filename := filepath.Base(mp4File)
			mp3Filename := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".mp3"
			outputPath := filepath.Join(outputDir, mp3Filename)

			jobs = append(jobs, ConversionJob{
				InputPath:  mp4File,
				OutputPath: outputPath,
				Status:     "pending",
			})
		}

		// 新しい変換開始メッセージタイプを返す
		return ConversionStartMsg{jobs: jobs}
	})
}

func (m Model) updateConversion(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ConversionStartMsg:
		m.state = StateConverting  // 状態遷移を追加
		m.jobs = msg.jobs
		m.current = 0
		m.completedCount = 0
		return m, tea.Batch(
			m.spinner.Tick,
			m.startParallelConversion(),
		)

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case JobDoneMsg:
		if msg.err != nil {
			m.jobs[msg.index].Status = "failed"
			m.jobs[msg.index].Error = msg.err
		} else {
			m.jobs[msg.index].Status = "completed"
		}
		m.completedCount++
		
		// 全て完了したかチェック
		if m.completedCount >= len(m.jobs) {
			m.state = StateDone
		}
		return m, nil

	case AllDoneMsg:
		m.state = StateDone
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

func (m Model) convertNext() tea.Cmd {
	if m.current >= len(m.jobs) {
		return func() tea.Msg { return AllDoneMsg{} }
	}

	job := m.jobs[m.current]
	return tea.Cmd(func() tea.Msg {
		err := convertMP4ToMP3(job.InputPath, job.OutputPath)
		return JobDoneMsg{index: m.current, err: err}
	})
}

// startParallelConversion 並列変換を開始
func (m Model) startParallelConversion() tea.Cmd {
	var cmds []tea.Cmd
	
	// セマフォを使用して並列数を制限
	sem := make(chan struct{}, m.maxConcurrent)
	
	for i, job := range m.jobs {
		index := i // クロージャ問題を回避
		currentJob := job
		
		cmds = append(cmds, tea.Cmd(func() tea.Msg {
			// セマフォを取得
			sem <- struct{}{}
			defer func() { <-sem }() // セマフォを解放
			
			// 変換実行
			err := convertMP4ToMP3(currentJob.InputPath, currentJob.OutputPath)
			return JobDoneMsg{index: index, err: err}
		}))
	}
	
	return tea.Batch(cmds...)
}

func (m Model) updateDone(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "enter", "esc":
			return m, tea.Quit
		case "r":
			// リスタート
			return initialModel(), nil
		}
	}
	return m, nil
}

func (m Model) View() string {
	switch m.state {
	case StateSelectingFolder:
		return m.folderSelectionView()
	case StateConfirmSelection:
		return m.confirmationView()
	case StateConverting:
		return m.conversionView()
	case StateDone:
		return m.doneView()
	}
	return ""
}

func (m Model) folderSelectionView() string {
	var s strings.Builder

	// 上部にマージンを追加
	s.WriteString("\n")
	s.WriteString(titleStyle.Render("🎵 MP4 to MP3 Converter"))
	s.WriteString("\n")
	s.WriteString(headerStyle.Render("Select a folder containing MP4 files:"))
	
	// 現在のパス表示を強化
	currentPath := m.filepicker.CurrentDirectory
	displayPath := currentPath
	if len(currentPath) > 70 {
		displayPath = "..." + currentPath[len(currentPath)-67:]
	}
	
	// パスの情報をより詳細に表示
	s.WriteString(fmt.Sprintf("\n📍 Current: %s", infoStyle.Render(displayPath)))
	
	// 実行時ディレクトリも表示
	if m.initialDir != currentPath {
		initialDisplayPath := m.initialDir
		if len(m.initialDir) > 50 {
			initialDisplayPath = "..." + m.initialDir[len(m.initialDir)-47:]
		}
		s.WriteString(fmt.Sprintf("\n🏠 Initial: %s", subtleStyle.Render(initialDisplayPath)))
	}
	
	// MP4ファイル数を事前表示
	if mp4Files, err := findMP4Files(currentPath); err == nil && len(mp4Files) > 0 {
		s.WriteString(fmt.Sprintf("\n🎬 MP4 files here: %s", successStyle.Render(fmt.Sprintf("%d", len(mp4Files)))))
	}
	
	s.WriteString("\n\n")

	s.WriteString(m.filepicker.View())
	s.WriteString("\n")

	// 拡張されたヘルプメッセージ
	helpText := strings.Join([]string{
		"Navigate: ↑/↓ • Enter folder: →/l/Enter • Back: ←/h • Select current: Space",
		"Shortcuts: ~ (Home) • / (Root) • . (Initial Dir) • p (Parent) • Quit: q/Ctrl+C",
	}, "\n")
	s.WriteString(helpStyle.Render(helpText))

	return s.String()
}

func (m Model) confirmationView() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("🎵 MP4 to MP3 Converter"))
	s.WriteString("\n\n")

	s.WriteString(fmt.Sprintf("📁 Selected folder: %s\n", infoStyle.Render(m.selectedPath)))
	s.WriteString(fmt.Sprintf("🎬 MP4 files found: %s\n\n", 
		successStyle.Render(fmt.Sprintf("%d", m.mp4Count))))

	if m.mp4Count == 0 {
		s.WriteString(errorStyle.Render("No MP4 files found in this folder."))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Press 'n' to go back and select another folder"))
	} else {
		s.WriteString("Conversion settings:\n")
		s.WriteString(subtleStyle.Render("• Bitrate: 64kbps (lightweight)\n"))
		s.WriteString(subtleStyle.Render("• Sample Rate: 22.05kHz\n"))
		s.WriteString(subtleStyle.Render("• Channels: Mono\n"))
		s.WriteString(subtleStyle.Render(fmt.Sprintf("• Output: %s/output/\n\n", m.selectedPath)))

		s.WriteString("Do you want to proceed with the conversion? ")
		s.WriteString(successStyle.Render("[y]es") + " / " + errorStyle.Render("[n]o"))
	}

	return s.String()
}

func (m Model) conversionView() string {
	if len(m.jobs) == 0 {
		return "Initializing conversion..."
	}

	var s strings.Builder

	s.WriteString(titleStyle.Render("🎵 Converting MP4 to MP3"))
	s.WriteString("\n\n")

	// 進捗表示（完了数ベース）
	percent := float64(m.completedCount) / float64(len(m.jobs))
	s.WriteString(fmt.Sprintf("Progress: %s %d/%d\n\n",
		m.progress.ViewAs(percent),
		m.completedCount,
		len(m.jobs)))

	// 並列処理中のメッセージ
	running := 0
	for _, job := range m.jobs {
		if job.Status == "pending" {
			running++
		}
	}
	
	if running > 0 {
		s.WriteString(fmt.Sprintf("%s Converting %d files in parallel...\n\n",
			m.spinner.View(),
			min(running, m.maxConcurrent)))
	}

	// 完了したジョブの表示
	for _, job := range m.jobs {
		filename := filepath.Base(job.InputPath)

		if job.Status == "completed" {
			s.WriteString(fmt.Sprintf("✓ %s\n", successStyle.Render(filename)))
		} else if job.Status == "failed" {
			s.WriteString(fmt.Sprintf("✗ %s (%s)\n",
				errorStyle.Render(filename),
				subtleStyle.Render(job.Error.Error())))
		}
	}

	s.WriteString("\n")
	s.WriteString(subtleStyle.Render("Press 'q' to quit"))

	return s.String()
}

// min ヘルパー関数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m Model) doneView() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("🎉 Conversion Complete!"))
	s.WriteString("\n\n")

	successCount := 0
	for _, job := range m.jobs {
		filename := filepath.Base(job.InputPath)
		if job.Status == "completed" {
			successCount++
			s.WriteString(fmt.Sprintf("✓ %s\n", successStyle.Render(filename)))
		} else if job.Status == "failed" {
			s.WriteString(fmt.Sprintf("✗ %s (%s)\n",
				errorStyle.Render(filename),
				subtleStyle.Render(job.Error.Error())))
		}
	}

	s.WriteString(fmt.Sprintf("\n%s\n\n",
		infoStyle.Render(fmt.Sprintf("Successfully converted %d/%d files",
			successCount, len(m.jobs)))))

	s.WriteString(fmt.Sprintf("📁 Output folder: %s\n\n",
		subtleStyle.Render(filepath.Join(m.selectedPath, "output"))))

	s.WriteString(helpStyle.Render("Press 'r' to restart • 'q' to quit"))

	return s.String()
}

// convertMP4ToMP3 MP4をMP3に変換
func convertMP4ToMP3(inputPath, outputPath string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-vn",           // 動画ストリームを無視
		"-acodec", "mp3", // MP3コーデック
		"-b:a", "64k",   // 64kbps（軽量）
		"-ar", "22050",  // 22.05kHz
		"-ac", "1",      // モノラル
		"-y",            // 上書き
		outputPath,
	)

	// FFmpegの出力を隠す
	cmd.Stdout = nil
	cmd.Stderr = nil

	return cmd.Run()
}

// checkFFmpeg FFmpegの存在確認
func checkFFmpeg() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

// findMP4Files MP4ファイルを検索
func findMP4Files(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".mp4" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// runInteractiveMode インタラクティブモードの実行
func runInteractiveMode() error {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running interactive mode: %w", err)
	}

	return nil
}

// runDirectMode ディレクト指定モードの実行
func runDirectMode(directory string) error {
	// FFmpegの確認
	if !checkFFmpeg() {
		return fmt.Errorf("ffmpeg not found. Please install ffmpeg first")
	}

	// 入力ディレクトリの確認
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", directory)
	}

	// MP4ファイルの検索
	mp4Files, err := findMP4Files(directory)
	if err != nil {
		return fmt.Errorf("error finding MP4 files: %w", err)
	}

	if len(mp4Files) == 0 {
		return fmt.Errorf("no MP4 files found in directory: %s", directory)
	}

	// outputディレクトリの作成
	outputDir := filepath.Join(directory, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	// ジョブの作成
	var jobs []ConversionJob
	for _, mp4File := range mp4Files {
		filename := filepath.Base(mp4File)
		mp3Filename := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".mp3"
		outputPath := filepath.Join(outputDir, mp3Filename)

		jobs = append(jobs, ConversionJob{
			InputPath:  mp4File,
			OutputPath: outputPath,
			Status:     "pending",
		})
	}

	// 変換モデルの作成
	model := Model{
		state:   StateConverting,
		jobs:    jobs,
		current: 0,
		progress: progress.New(
			progress.WithDefaultGradient(),
			progress.WithWidth(40),
			progress.WithoutPercentage(),
		),
		spinner: func() spinner.Model {
			s := spinner.New()
			s.Spinner = spinner.Dot
			s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			return s
		}(),
	}

	// Bubble Tea プログラムの実行
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mp4converter [directory]",
		Short: "Convert MP4 files to MP3 (optimized for meeting recordings)",
		Long: `A beautiful interactive CLI tool to convert MP4 files to MP3 format.

If no directory is specified, an interactive folder picker will be shown.
Optimized for meeting recordings with lightweight settings:
- Bitrate: 64kbps
- Sample Rate: 22.05kHz  
- Channels: Mono`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				// インタラクティブモード
				return runInteractiveMode()
			} else {
				// ディレクト指定モード
				return runDirectMode(args[0])
			}
		},
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("mp4converter v2.0.0")
			fmt.Println("Built with ❤️  using Charm")
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("%s %s\n", errorStyle.Render("Error:"), err.Error())
		os.Exit(1)
	}
}