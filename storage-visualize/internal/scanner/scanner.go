package scanner

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/user/storage-visualizer/internal/model"
)

// Scanner はファイルシステムをスキャンするための構造体
type Scanner struct {
	// スキャン中のエラーを格納
	Error error
	// 処理中のディレクトリ数
	processingCount int
	// 同期用のミューテックス
	mutex sync.Mutex
}

// NewScanner は新しいScannerを作成する
func NewScanner() *Scanner {
	return &Scanner{}
}

// Scan は指定されたパスからスキャンを開始する
func (s *Scanner) Scan(path string) (*model.DirNode, error) {
	// パスの存在確認
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// ルートノードの作成
	root := model.NewDirNode(filepath.Base(path), path, info.IsDir())

	// ディレクトリでない場合はサイズを設定して終了
	if !info.IsDir() {
		root.Size = info.Size()
		return root, nil
	}

	// ディレクトリの場合は再帰的にスキャン
	err = s.scanDir(root)
	if err != nil {
		return nil, err
	}

	// 相対サイズの計算
	root.CalculateRelativeSizes()

	return root, nil
}

// scanDir はディレクトリを再帰的にスキャンする
func (s *Scanner) scanDir(node *model.DirNode) error {
	// ディレクトリ内のファイル一覧を取得
	entries, err := os.ReadDir(node.AbsolutePath)
	if err != nil {
		return err
	}

	// 各エントリを処理
	var totalSize int64
	var wg sync.WaitGroup
	errChan := make(chan error, len(entries))

	for _, entry := range entries {
		// 隠しファイルはスキップ（オプション）
		if entry.Name()[0] == '.' {
			continue
		}

		// エントリの絶対パスを取得
		entryPath := filepath.Join(node.AbsolutePath, entry.Name())

		// ファイル情報を取得
		info, err := entry.Info()
		if err != nil {
			// 権限エラーなどは無視して続行
			continue
		}

		// 新しいノードを作成
		child := model.NewDirNode(entry.Name(), entryPath, info.IsDir())
		node.AddChild(child)

		if info.IsDir() {
			// ディレクトリの場合は非同期で処理
			wg.Add(1)
			s.mutex.Lock()
			s.processingCount++
			s.mutex.Unlock()

			go func(dirNode *model.DirNode) {
				defer wg.Done()
				defer func() {
					s.mutex.Lock()
					s.processingCount--
					s.mutex.Unlock()
				}()

				err := s.scanDir(dirNode)
				if err != nil {
					errChan <- err
				}
			}(child)
		} else {
			// ファイルの場合はサイズを設定
			child.Size = info.Size()
			totalSize += info.Size()
		}
	}

	// 非同期処理の完了を待つ
	wg.Wait()
	close(errChan)

	// エラーチェック
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	// ディレクトリのサイズを計算（子ノードの合計）
	for _, child := range node.Children {
		totalSize += child.Size
	}
	node.Size = totalSize

	return nil
}

// GetProcessingCount は現在処理中のディレクトリ数を返す
func (s *Scanner) GetProcessingCount() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.processingCount
}
