# MP4 to MP3 Converter (GUI版)

🎵 MP4ファイルをMP3に変換する直感的なデスクトップアプリケーション。会議録音などの音声用途に最適化されています。

## ✨ 特徴

- **直感的なGUI**: 美しく使いやすいデスクトップインターフェース
- **フォルダ選択**: ワンクリックでフォルダを選択
- **一括変換**: 複数のMP4ファイルを同時に変換
- **リアルタイム進捗**: ファイルごとと全体の進捗を表示
- **設定カスタマイズ**: ビットレート、サンプリングレート、チャンネル数を調整可能
- **ログ表示**: 変換状況とエラーを詳細に表示

## 🎯 最適化設定

会議録音や音声コンテンツに特化したデフォルト設定：

- **ビットレート**: 64kbps（軽量）
- **サンプリングレート**: 22.05kHz
- **チャンネル**: モノラル
- **フォーマット**: MP3

## 📦 インストール

### 必要なもの

1. **Go 1.19+** - [Goをダウンロード](https://golang.org/dl/)
2. **Node.js 16+** - [Node.jsをダウンロード](https://nodejs.org/)
3. **Wails v2** - [Wailsをインストール](https://wails.io/docs/gettingstarted/installation)
4. **FFmpeg** - 音声変換に必要

#### FFmpegのインストール

**macOS (Homebrew):**
```bash
brew install ffmpeg
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install ffmpeg
```

**Windows (Chocolatey):**
```bash
choco install ffmpeg
```

#### Wailsのインストール

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### ビルド

1. リポジトリをクローン:
```bash
git clone <repository-url>
cd mp4Tomp3Converter/gui/mp4ToMp3Gui
```

2. 依存関係をインストール:
```bash
wails build
```

### 開発

開発モードで実行:
```bash
wails dev
```

## 🚀 使い方

### 1. フォルダ選択
- 「フォルダを選択」ボタンをクリック
- MP4ファイルが含まれるフォルダを選択
- 自動的にMP4ファイルが検出されます

### 2. 変換設定
以下の設定を必要に応じて調整：
- **ビットレート**: 64kbps（軽量）～ 320kbps（最高品質）
- **サンプリングレート**: 22.05kHz（軽量）～ 48kHz（高品質）
- **チャンネル**: モノラル または ステレオ

### 3. 変換実行
- 「変換を開始」ボタンをクリック
- 進捗がリアルタイムで表示されます
- ログでエラーや完了状況を確認できます

## 📁 出力構造

選択したディレクトリに `output` フォルダが作成されます：

```
選択したディレクトリ/
├── video1.mp4
├── video2.mp4
└── output/
    ├── video1.mp3
    └── video2.mp3
```

## 🎨 UI概要

### メイン画面
- **ヘッダー**: アプリケーション名と説明
- **フォルダ選択**: 変換対象フォルダの選択
- **変換設定**: 音声品質の調整
- **変換ボタン**: 一括変換の開始

### 進捗表示
- **全体進捗バー**: 全ファイルの変換進捗
- **ファイル別進捗**: 各ファイルの個別進捗
- **ステータス表示**: 待機中/変換中/完了/エラー

### ログ画面
- **タイムスタンプ付きログ**: 詳細な実行履歴
- **エラー表示**: 問題発生時の詳細情報

## 🐛 トラブルシューティング

**FFmpegが見つからない:**
- FFmpegがインストールされていることを確認
- PATHにFFmpegが含まれていることを確認

**フォルダが選択できない:**
- フォルダの読み取り権限を確認
- 管理者権限で実行してみる

**変換が失敗する:**
- ログを確認してエラーの詳細を確認
- MP4ファイルが破損していないか確認
- ディスクの空き容量を確認

## 🔧 開発者向け

### プロジェクト構造
```
gui/mp4ToMp3Gui/
├── app.go              # Goバックエンド
├── main.go             # アプリケーションエントリーポイント
├── wails.json          # Wails設定
├── frontend/           # Reactフロントエンド
│   ├── src/
│   │   ├── App.tsx     # メインコンポーネント
│   │   └── App.css     # スタイルシート
│   └── wailsjs/        # 自動生成バインディング
└── build/              # ビルド成果物
```

### バックエンド機能
- `SelectFolder()`: フォルダ選択ダイアログ
- `GetMP4Files(directory)`: MP4ファイル検索
- `ConvertFiles(files, settings)`: 変換実行

### フロントエンド機能
- React + TypeScript
- レスポンシブデザイン
- リアルタイム進捗表示
- エラーハンドリング

Built with ❤️ using [Wails](https://wails.io/) and [React](https://reactjs.org/)
