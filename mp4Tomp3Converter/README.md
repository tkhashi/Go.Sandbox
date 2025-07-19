# MP4 to MP3 Converter

## 概要
このツールは、MP4ファイルをMP3形式に変換するためのCLIアプリケーションです。特に会議録音の変換に最適化されています。

## 特徴
- **軽量な設定**
  - ビットレート: 64kbps
  - サンプルレート: 22.05kHz
  - チャンネル: モノラル
- **美しいCLIインターフェース**
  - [Bubble Tea](https://github.com/charmbracelet/bubbletea)を使用した直感的なTUI
- **簡単な操作**
  - MP4ファイルを指定するだけで変換可能

## 必要条件
- [FFmpeg](https://ffmpeg.org/) がインストールされていること
- Go 1.18以上

## インストール
1. このリポジトリをクローンします。
   ```bash
   git clone https://github.com/your-repo/mp4Tomp3Converter.git
   ```
2. 必要な依存関係をインストールします。
   ```bash
   go mod tidy
   ```

## 使用方法
1. 変換したいMP4ファイルが含まれるディレクトリを指定して実行します。
   ```bash
   go run mp4_converter.go [directory]
   ```
   例:
   ```bash
   go run mp4_converter.go ./videos
   ```
2. 変換されたMP3ファイルは、指定したディレクトリ内の`output`フォルダに保存されます。

## コマンド
- **変換**
  ```bash
  mp4converter [directory]
  ```
  指定したディレクトリ内のMP4ファイルをMP3に変換します。

- **バージョン情報**
  ```bash
  mp4converter version
  ```
  ツールのバージョン情報を表示します。

## 注意事項
- FFmpegがインストールされていない場合、ツールは動作しません。インストール方法については[こちら](https://ffmpeg.org/download.html)を参照してください。
- 入力ディレクトリにMP4ファイルが存在しない場合、エラーが発生します。

## FFmpegのインストール方法

このツールを使用するには、FFmpegが必要です。以下の手順でインストールしてください。

### Homebrewを使用する場合
1. Homebrewがインストールされていない場合は、[公式サイト](https://brew.sh/)を参照してインストールしてください。
2. 以下のコマンドを実行してFFmpegをインストールします。
   ```bash
   brew install ffmpeg
   ```

### curlを使用する場合
1. 以下のコマンドを実行してFFmpegをダウンロードします。
   ```bash
   curl -L -o ffmpeg.zip https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip
   ```
2. ダウンロードしたZIPファイルを解凍します。
   ```bash
   unzip ffmpeg.zip
   ```
3. 解凍したフォルダ内の`bin`ディレクトリにある`ffmpeg`をパスに追加します。
   ```bash
   export PATH=$PATH:/path/to/ffmpeg/bin
   ```
   ※ `/path/to/ffmpeg/bin`は解凍したフォルダのパスに置き換えてください。

インストール後、以下のコマンドでFFmpegが正しくインストールされているか確認できます。
```bash
ffmpeg -version
```

## ライセンス
このプロジェクトはMITライセンスの下で提供されています。
