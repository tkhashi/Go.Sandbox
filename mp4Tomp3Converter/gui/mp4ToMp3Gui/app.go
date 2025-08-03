package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed resources/bin/*
var resourceFS embed.FS

// App struct
type App struct {
	ctx            context.Context
	selectedFolder string // 選択されたフォルダのパスを保存
	ffmpegPath     string // FFmpegのパス
}

// ConversionProgress 変換進捗の情報
type ConversionProgress struct {
	FileName      string  `json:"fileName"`
	Status        string  `json:"status"`        // "pending", "converting", "completed", "failed"
	Progress      int     `json:"progress"`      // 0-100
	Error         string  `json:"error,omitempty"`
	OverallIndex  int     `json:"overallIndex"`  // 現在のファイルのインデックス
	TotalFiles    int     `json:"totalFiles"`    // 総ファイル数
}

// ConversionSettings 変換設定
type ConversionSettings struct {
	Bitrate    string `json:"bitrate"`
	SampleRate string `json:"sampleRate"`
	Channels   string `json:"channels"`
}

// ConversionJob 変換ジョブ
type ConversionJob struct {
	InputPath  string
	OutputPath string
	Status     string
	Error      error
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// FFmpegバイナリを初期化
	if err := a.initFFmpeg(); err != nil {
		fmt.Printf("FFmpeg initialization error: %v\n", err)
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// SelectFolder opens a folder selection dialog
func (a *App) SelectFolder() (string, error) {
	// Wailsのランタイムを使用してフォルダ選択ダイアログを開く
	options := runtime.OpenDialogOptions{
		Title: "MP4ファイルが含まれるフォルダを選択",
	}
	
	selectedPath, err := runtime.OpenDirectoryDialog(a.ctx, options)
	if err != nil {
		return "", fmt.Errorf("フォルダ選択エラー: %v", err)
	}
	
	if selectedPath == "" {
		return "", fmt.Errorf("フォルダが選択されませんでした")
	}
	
	// 選択されたフォルダパスを保存
	a.selectedFolder = selectedPath
	
	return selectedPath, nil
}

// GetMP4Files returns a list of MP4 files in the given directory
func (a *App) GetMP4Files(directory string) ([]string, error) {
	if directory == "" {
		return nil, fmt.Errorf("ディレクトリが指定されていません")
	}
	
	var mp4Files []string
	
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".mp4") {
			// 相対パスではなく、ファイル名のみを返す
			mp4Files = append(mp4Files, info.Name())
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("MP4ファイルの検索エラー: %v", err)
	}
	
	return mp4Files, nil
}

// CheckFFmpeg FFmpegの存在確認
func (a *App) CheckFFmpeg() bool {
	// 組み込みFFmpegがあるかチェック
	if a.ffmpegPath != "" {
		if _, err := os.Stat(a.ffmpegPath); err == nil {
			return true
		}
	}
	
	// システムのFFmpegをチェック
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

// initFFmpeg FFmpegバイナリを初期化
func (a *App) initFFmpeg() error {
	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "mp4tomp3-ffmpeg-*")
	if err != nil {
		return fmt.Errorf("一時ディレクトリの作成に失敗: %v", err)
	}
	
	// プラットフォームに応じてバイナリ名とリソースパスを決定
	var binaryName, resourcePath string
	switch goruntime.GOOS {
	case "darwin":
		binaryName = "ffmpeg"
		if goruntime.GOARCH == "arm64" {
			resourcePath = "resources/bin/ffmpeg"  // 現在のファイル
		} else {
			resourcePath = "resources/bin/ffmpeg-darwin-amd64"
		}
	case "windows":
		binaryName = "ffmpeg.exe"
		resourcePath = "resources/bin/ffmpeg-windows-amd64.exe"
	case "linux":
		binaryName = "ffmpeg"
		resourcePath = "resources/bin/ffmpeg-linux-amd64"
	default:
		return fmt.Errorf("サポートされていないプラットフォーム: %s", goruntime.GOOS)
	}
	
	// 埋め込まれたFFmpegバイナリを読み込み
	ffmpegData, err := resourceFS.ReadFile(resourcePath)
	if err != nil {
		// 埋め込みFFmpegが見つからない場合はシステムのFFmpegを使用
		fmt.Printf("Embedded FFmpeg not found (%s), will use system FFmpeg: %v\n", resourcePath, err)
		return nil
	}
	
	// 一時ディレクトリにFFmpegバイナリを書き込み
	ffmpegPath := filepath.Join(tempDir, binaryName)
	err = os.WriteFile(ffmpegPath, ffmpegData, 0755)
	if err != nil {
		return fmt.Errorf("FFmpegバイナリの書き込みに失敗: %v", err)
	}
	
	a.ffmpegPath = ffmpegPath
	fmt.Printf("FFmpeg extracted to: %s\n", ffmpegPath)
	
	return nil
}

// getFFmpegCommand FFmpegコマンドのパスを取得
func (a *App) getFFmpegCommand() string {
	if a.ffmpegPath != "" {
		return a.ffmpegPath
	}
	return "ffmpeg" // システムのFFmpegを使用
}

// ConvertFiles starts the conversion process
func (a *App) ConvertFiles(files []string, settings ConversionSettings) error {
	if !a.CheckFFmpeg() {
		return fmt.Errorf("FFmpegが見つかりません。FFmpegをインストールしてください")
	}
	
	if len(files) == 0 {
		return fmt.Errorf("変換するファイルがありません")
	}
	
	if a.selectedFolder == "" {
		return fmt.Errorf("フォルダが選択されていません")
	}
	
	// 出力ディレクトリの作成（選択されたフォルダ内のoutputフォルダ）
	outputDir := filepath.Join(a.selectedFolder, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("出力ディレクトリの作成に失敗: %v", err)
	}
	
	// 変換を非同期で開始
	go a.convertFilesAsync(files, settings, outputDir)
	
	return nil
}

// convertFilesAsync 非同期で変換を実行
func (a *App) convertFilesAsync(files []string, settings ConversionSettings, outputDir string) {
	// 並列処理用のセマフォ（最大3つまで同時実行）
	maxConcurrent := 3
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	
	// エラー用のスライス
	var errors []error
	var errorsMutex sync.Mutex
	
	for i, file := range files {
		wg.Add(1)
		sem <- struct{}{} // セマフォを取得
		
		go func(filename string, index int) {
			defer wg.Done()
			defer func() { <-sem }() // セマフォを解放
			
			// 変換開始を通知
			a.emitProgress(ConversionProgress{
				FileName:     filename,
				Status:       "converting",
				Progress:     0,
				OverallIndex: index,
				TotalFiles:   len(files),
			})
			
			inputPath := filepath.Join(a.selectedFolder, filename)
			outputPath := filepath.Join(outputDir, strings.Replace(filename, ".mp4", ".mp3", 1))
			
			err := a.convertMP4ToMP3WithProgress(inputPath, outputPath, settings, filename, index, len(files))
			if err != nil {
				errorsMutex.Lock()
				errors = append(errors, fmt.Errorf("%s: %v", filename, err))
				errorsMutex.Unlock()
				
				// エラーを通知
				a.emitProgress(ConversionProgress{
					FileName:     filename,
					Status:       "failed",
					Progress:     0,
					Error:        err.Error(),
					OverallIndex: index,
					TotalFiles:   len(files),
				})
			} else {
				// 完了を通知
				a.emitProgress(ConversionProgress{
					FileName:     filename,
					Status:       "completed",
					Progress:     100,
					OverallIndex: index,
					TotalFiles:   len(files),
				})
			}
		}(file, i)
	}
	
	wg.Wait()
	
	// 全体完了を通知
	a.emitProgress(ConversionProgress{
		FileName:     "all",
		Status:       "all_completed",
		Progress:     100,
		OverallIndex: len(files),
		TotalFiles:   len(files),
	})
}

// convertMP4ToMP3WithProgress MP4をMP3に変換（進捗付き）
func (a *App) convertMP4ToMP3WithProgress(inputPath, outputPath string, settings ConversionSettings, fileName string, index, total int) error {
	ffmpegCmd := a.getFFmpegCommand()
	
	cmd := exec.Command(ffmpegCmd,
		"-i", inputPath,
		"-vn",              // 動画ストリームを無視
		"-acodec", "mp3",   // MP3コーデック
		"-b:a", settings.Bitrate,     // ビットレート
		"-ar", settings.SampleRate,   // サンプリングレート
		"-ac", settings.Channels,     // チャンネル数
		"-y",               // 上書き
		outputPath,
	)
	
	// 進捗をシミュレート（FFmpegの実際の進捗は取得困難のため）
	go func() {
		for progress := 10; progress < 100; progress += 30 {
			a.emitProgress(ConversionProgress{
				FileName:     fileName,
				Status:       "converting",
				Progress:     progress,
				OverallIndex: index,
				TotalFiles:   total,
			})
			// 変換時間をシミュレート
			time.Sleep(500 * time.Millisecond)
		}
	}()
	
	// FFmpegの出力を隠す
	cmd.Stdout = nil
	cmd.Stderr = nil
	
	return cmd.Run()
}

// emitProgress 進捗をフロントエンドに送信
func (a *App) emitProgress(progress ConversionProgress) {
	runtime.EventsEmit(a.ctx, "conversion-progress", progress)
}

// convertMP4ToMP3 MP4をMP3に変換（TUIアプリから移植）
func (a *App) convertMP4ToMP3(inputPath, outputPath string, settings ConversionSettings) error {
	ffmpegCmd := a.getFFmpegCommand()
	
	cmd := exec.Command(ffmpegCmd,
		"-i", inputPath,
		"-vn",              // 動画ストリームを無視
		"-acodec", "mp3",   // MP3コーデック
		"-b:a", settings.Bitrate,     // ビットレート
		"-ar", settings.SampleRate,   // サンプリングレート
		"-ac", settings.Channels,     // チャンネル数
		"-y",               // 上書き
		outputPath,
	)
	
	// FFmpegの出力を隠す
	cmd.Stdout = nil
	cmd.Stderr = nil
	
	return cmd.Run()
}

// Cleanup アプリケーション終了時のクリーンアップ
func (a *App) Cleanup(ctx context.Context) {
	if a.ffmpegPath != "" {
		// 一時ディレクトリを削除
		tempDir := filepath.Dir(a.ffmpegPath)
		os.RemoveAll(tempDir)
	}
}
