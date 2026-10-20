import {useState, useEffect} from 'react';
import './App.css';
import {SelectFolder, GetMP4Files, ConvertFiles, CheckFFmpeg} from "../wailsjs/go/main/App";
import {main} from "../wailsjs/go/models";
import {EventsOn} from "../wailsjs/runtime/runtime";

// 変換設定の型定義
interface ConversionSettings {
    bitrate: string;
    sampleRate: string;
    channels: string;
}

// ファイルの変換状況の型定義
interface ConversionFile {
    name: string;
    status: 'pending' | 'converting' | 'completed' | 'failed';
    progress: number;
    error?: string;
}

// 進捗データの型定義
interface ConversionProgress {
    fileName: string;
    status: string;
    progress: number;
    error?: string;
    overallIndex: number;
    totalFiles: number;
}

function App() {
    const [selectedFolder, setSelectedFolder] = useState<string>('');
    const [mp4Files, setMp4Files] = useState<string[]>([]);
    const [conversionSettings, setConversionSettings] = useState<ConversionSettings>({
        bitrate: '64k',
        sampleRate: '22050',
        channels: '1'
    });
    const [isConverting, setIsConverting] = useState(false);
    const [conversionFiles, setConversionFiles] = useState<ConversionFile[]>([]);
    const [overallProgress, setOverallProgress] = useState(0);
    const [logs, setLogs] = useState<string[]>([]);
    const [isDragging, setIsDragging] = useState(false);
    const [ffmpegAvailable, setFfmpegAvailable] = useState<boolean>(true);

    // FFmpegの確認
    useEffect(() => {
        CheckFFmpeg().then(setFfmpegAvailable).catch(() => setFfmpegAvailable(false));
    }, []);

    // 変換進捗のイベントリスナー
    useEffect(() => {
        const unsubscribe = EventsOn("conversion-progress", (data: ConversionProgress) => {
            handleConversionProgress(data);
        });

        return () => {
            if (unsubscribe) unsubscribe();
        };
    }, []);

    // フォルダ選択
    const selectFolder = async () => {
        try {
            const folderPath = await SelectFolder();
            if (!folderPath) {
                addLog("フォルダの選択がキャンセルされました");
                return;
            }
            
            const files = await GetMP4Files(folderPath);
            
            setSelectedFolder(folderPath);
            setMp4Files(files);
            addLog(`フォルダを選択しました: ${folderPath}`);
            addLog(`${files.length}個のMP4ファイルが見つかりました`);
        } catch (error) {
            addLog(`エラー: ${error}`);
        }
    };

    // 変換開始
    const startConversion = async () => {
        if (!ffmpegAvailable) {
            addLog("FFmpegが見つかりません。FFmpegをインストールしてから再試行してください");
            return;
        }
        
        if (mp4Files.length === 0) {
            addLog("変換するファイルがありません");
            return;
        }

        setIsConverting(true);
        setOverallProgress(0);
        
        // ファイルリストを初期化
        const files: ConversionFile[] = mp4Files.map(file => ({
            name: file,
            status: 'pending',
            progress: 0
        }));
        setConversionFiles(files);
        
        addLog("変換を開始します...");
        
        try {
            // バックエンドの ConversionSettings 型に合わせて変換
            const settings = new main.ConversionSettings({
                bitrate: conversionSettings.bitrate,
                sampleRate: conversionSettings.sampleRate,
                channels: conversionSettings.channels
            });
            
            // 非同期変換処理を開始（進捗はイベントで受信）
            await ConvertFiles(mp4Files, settings);
            
        } catch (error) {
            addLog(`変換エラー: ${error}`);
            setConversionFiles(prev => 
                prev.map(f => ({ ...f, status: 'failed' as const }))
            );
            setIsConverting(false);
        }
    };

    // ログを追加
    const addLog = (message: string) => {
        const timestamp = new Date().toLocaleTimeString();
        setLogs(prev => [...prev, `[${timestamp}] ${message}`]);
    };

    // 変換進捗を処理
    const handleConversionProgress = (data: ConversionProgress) => {
        if (data.status === "all_completed") {
            // 全体完了
            setIsConverting(false);
            setOverallProgress(100);
            addLog("すべての変換が完了しました！");
            return;
        }

        // 個別ファイルの進捗を更新
        setConversionFiles(prev => 
            prev.map(file => 
                file.name === data.fileName 
                    ? { 
                        ...file, 
                        status: data.status as 'pending' | 'converting' | 'completed' | 'failed',
                        progress: data.progress,
                        error: data.error 
                    }
                    : file
            )
        );

        // 全体進捗を更新
        const progressPerFile = 100 / data.totalFiles;
        const completedFiles = data.overallIndex;
        const currentFileProgress = data.progress / 100 * progressPerFile;
        const totalProgress = (completedFiles * progressPerFile) + currentFileProgress;
        setOverallProgress(Math.min(totalProgress, 100));

        // ログを追加
        if (data.status === "converting" && data.progress === 0) {
            addLog(`${data.fileName} の変換を開始`);
        } else if (data.status === "completed") {
            addLog(`${data.fileName} の変換が完了しました`);
        } else if (data.status === "failed") {
            addLog(`${data.fileName} の変換に失敗: ${data.error}`);
        }
    };

    // ドラッグ&ドロップのハンドラ
    const handleDragOver = (e: React.DragEvent) => {
        e.preventDefault();
        setIsDragging(true);
    };

    const handleDragLeave = (e: React.DragEvent) => {
        e.preventDefault();
        setIsDragging(false);
    };

    const handleDrop = (e: React.DragEvent) => {
        e.preventDefault();
        setIsDragging(false);
        
        if (isConverting) {
            addLog("変換中はファイルをドロップできません");
            return;
        }

        const files = Array.from(e.dataTransfer.files);
        const mp4FilesList = files
            .filter(file => file.name.toLowerCase().endsWith('.mp4'))
            .map(file => file.name);

        if (mp4FilesList.length === 0) {
            addLog("MP4ファイルが含まれていません");
            return;
        }

        setMp4Files(mp4FilesList);
        setSelectedFolder(''); // フォルダ選択をクリア
        addLog(`${mp4FilesList.length}個のMP4ファイルがドロップされました`);
        mp4FilesList.forEach(file => addLog(`- ${file}`));
    };

    return (
        <div id="App" 
             onDragOver={handleDragOver}
             onDragLeave={handleDragLeave}
             onDrop={handleDrop}
             className={isDragging ? 'dragging' : ''}
        >
            <div className="container">
                {/* ドロップオーバーレイ */}
                {isDragging && (
                    <div className="drop-overlay">
                        <div className="drop-message">
                            <h2>📎 MP4ファイルをドロップしてください</h2>
                            <p>複数ファイルの同時ドロップが可能です</p>
                        </div>
                    </div>
                )}

                {/* ヘッダー */}
                <header className="header">
                    <h1>🎵 MP4 to MP3 Converter</h1>
                    <p>複数のMP4ファイルをMP3形式に一括変換</p>
                    {!ffmpegAvailable && (
                        <div className="ffmpeg-warning">
                            ⚠️ FFmpegが見つかりません。変換を行うにはFFmpegをインストールしてください。
                        </div>
                    )}
                </header>

                {/* フォルダ選択セクション */}
                <section className="section">
                    <h2>📁 ファイル選択</h2>
                    <div className="folder-selection">
                        <div className="selected-folder">
                            {selectedFolder ? (
                                <span className="folder-path">{selectedFolder}</span>
                            ) : mp4Files.length > 0 ? (
                                <span className="folder-path">ドロップされたファイル ({mp4Files.length}個)</span>
                            ) : (
                                <span className="placeholder">フォルダを選択するか、MP4ファイルをドロップしてください</span>
                            )}
                        </div>
                        <button 
                            className="btn btn-primary" 
                            onClick={selectFolder}
                            disabled={isConverting}
                        >
                            フォルダを選択
                        </button>
                    </div>
                    {mp4Files.length > 0 && (
                        <div className="file-count">
                            <span className="count">{mp4Files.length}個のMP4ファイルが見つかりました</span>
                            <div className="file-list">
                                {mp4Files.map((file, index) => (
                                    <div key={index} className="file-item">
                                        📄 {file}
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}
                </section>

                {/* 変換設定セクション */}
                <section className="section">
                    <h2>⚙️ 変換設定</h2>
                    <div className="settings-grid">
                        <div className="setting-item">
                            <label>ビットレート:</label>
                            <select 
                                value={conversionSettings.bitrate} 
                                onChange={(e) => setConversionSettings(prev => ({...prev, bitrate: e.target.value}))}
                                disabled={isConverting}
                            >
                                <option value="64k">64kbps (軽量)</option>
                                <option value="128k">128kbps (標準)</option>
                                <option value="192k">192kbps (高品質)</option>
                                <option value="320k">320kbps (最高品質)</option>
                            </select>
                        </div>
                        <div className="setting-item">
                            <label>サンプリングレート:</label>
                            <select 
                                value={conversionSettings.sampleRate} 
                                onChange={(e) => setConversionSettings(prev => ({...prev, sampleRate: e.target.value}))}
                                disabled={isConverting}
                            >
                                <option value="22050">22.05kHz (軽量)</option>
                                <option value="44100">44.1kHz (CD品質)</option>
                                <option value="48000">48kHz (高品質)</option>
                            </select>
                        </div>
                        <div className="setting-item">
                            <label>チャンネル:</label>
                            <select 
                                value={conversionSettings.channels} 
                                onChange={(e) => setConversionSettings(prev => ({...prev, channels: e.target.value}))}
                                disabled={isConverting}
                            >
                                <option value="1">モノラル</option>
                                <option value="2">ステレオ</option>
                            </select>
                        </div>
                    </div>
                </section>

                {/* 変換開始ボタン */}
                <section className="section">
                    <button 
                        className="btn btn-convert" 
                        onClick={startConversion}
                        disabled={mp4Files.length === 0 || isConverting || !ffmpegAvailable}
                    >
                        {!ffmpegAvailable ? 'FFmpegが必要です' : 
                         isConverting ? '変換中...' : '変換を開始'}
                    </button>
                </section>

                {/* 進捗表示セクション */}
                {isConverting && (
                    <section className="section">
                        <h2>📊 変換進捗</h2>
                        <div className="progress-section">
                            <div className="overall-progress">
                                <label>全体進捗:</label>
                                <div className="progress-bar">
                                    <div 
                                        className="progress-fill" 
                                        style={{width: `${overallProgress}%`}}
                                    ></div>
                                </div>
                                <span className="progress-text">{Math.round(overallProgress)}%</span>
                            </div>
                            
                            <div className="file-progress-list">
                                {conversionFiles.map((file, index) => (
                                    <div key={index} className="file-progress-item">
                                        <div className="file-info">
                                            <span className="file-name">{file.name}</span>
                                            <span className={`status status-${file.status}`}>
                                                {file.status === 'pending' && '待機中'}
                                                {file.status === 'converting' && '変換中'}
                                                {file.status === 'completed' && '完了'}
                                                {file.status === 'failed' && 'エラー'}
                                            </span>
                                        </div>
                                        <div className="progress-bar file-progress">
                                            <div 
                                                className="progress-fill" 
                                                style={{width: `${file.progress}%`}}
                                            ></div>
                                        </div>
                                        <span className="progress-text">{file.progress}%</span>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </section>
                )}

                {/* ログセクション */}
                <section className="section">
                    <h2>📝 ログ</h2>
                    <div className="log-container">
                        {logs.length === 0 ? (
                            <div className="log-empty">ログはありません</div>
                        ) : (
                            logs.map((log, index) => (
                                <div key={index} className="log-item">{log}</div>
                            ))
                        )}
                    </div>
                </section>
            </div>
        </div>
    )
}

export default App
