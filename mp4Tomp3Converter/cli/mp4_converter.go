package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
)

// ConversionJob 変換ジョブ
type ConversionJob struct {
	InputPath  string
	OutputPath string
	Status     string
	Error      error
}

// Model Bubble Tea モデル
type Model struct {
	jobs     []ConversionJob
	current  int
	progress progress.Model
	spinner  spinner.Model
	done     bool
	err      error
}

// メッセージ定義
type JobDoneMsg struct {
	index int
	err   error
}

type AllDoneMsg struct{}

func initialModel(jobs []ConversionJob) Model {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		jobs:     jobs,
		progress: p,
		spinner:  s,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.convertNext(),
	)
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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
		m.current++
		return m, m.convertNext()

	case AllDoneMsg:
		m.done = true
		return m, tea.Quit

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

func (m Model) View() string {
	if m.done {
		return m.finalView()
	}

	var s strings.Builder

	s.WriteString(titleStyle.Render("🎵 MP4 to MP3 Converter"))
	s.WriteString("\n\n")

	// 進捗表示
	percent := float64(m.current) / float64(len(m.jobs))
	s.WriteString(fmt.Sprintf("Progress: %s %d/%d\n\n",
		m.progress.ViewAs(percent),
		m.current,
		len(m.jobs)))

	// 現在の処理
	if m.current < len(m.jobs) {
		currentFile := filepath.Base(m.jobs[m.current].InputPath)
		s.WriteString(fmt.Sprintf("%s Converting: %s\n\n",
			m.spinner.View(),
			infoStyle.Render(currentFile)))
	}

	// 完了したジョブ
	for i := 0; i < m.current && i < len(m.jobs); i++ {
		job := m.jobs[i]
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

func (m Model) finalView() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("🎵 Conversion Complete!"))
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

	s.WriteString(fmt.Sprintf("\n%s\n",
		infoStyle.Render(fmt.Sprintf("Successfully converted %d/%d files",
			successCount, len(m.jobs)))))

	return s.String()
}

// convertMP4ToMP3 MP4をMP3に変換
func convertMP4ToMP3(inputPath, outputPath string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-vn",            // 動画ストリームを無視
		"-acodec", "mp3", // MP3コーデック
		"-b:a", "64k", // 64kbps（軽量）
		"-ar", "22050", // 22.05kHz
		"-ac", "1", // モノラル
		"-y", // 上書き
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

// runConversion 変換処理の実行
func runConversion(directory string) error {
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

	// Bubble Tea プログラムの実行
	p := tea.NewProgram(initialModel(jobs))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mp4converter [directory]",
		Short: "Convert MP4 files to MP3 (optimized for meeting recordings)",
		Long: `A beautiful CLI tool to convert MP4 files to MP3 format.
Optimized for meeting recordings with lightweight settings:
- Bitrate: 64kbps
- Sample Rate: 22.05kHz  
- Channels: Mono`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConversion(args[0])
		},
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("mp4converter v1.0.0")
			fmt.Println("Built with ❤️  using Charm")
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("%s %s\n", errorStyle.Render("Error:"), err.Error())
		os.Exit(1)
	}
}
