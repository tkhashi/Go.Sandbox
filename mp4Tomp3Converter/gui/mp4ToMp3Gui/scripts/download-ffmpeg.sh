#!/bin/bash

# FFmpegバイナリをダウンロードするスクリプト

set -e

RESOURCES_DIR="$(dirname "$0")/../resources/bin"
mkdir -p "$RESOURCES_DIR"

echo "Downloading FFmpeg binaries..."

# macOS ARM64 (already downloaded)
if [ ! -f "$RESOURCES_DIR/ffmpeg-darwin-arm64" ]; then
    echo "Downloading FFmpeg for macOS ARM64..."
    curl -L -o "$RESOURCES_DIR/ffmpeg-darwin-arm64.zip" "https://evermeet.cx/ffmpeg/getrelease/zip"
    cd "$RESOURCES_DIR"
    unzip ffmpeg-darwin-arm64.zip
    mv ffmpeg ffmpeg-darwin-arm64
    rm ffmpeg-darwin-arm64.zip
    cd - > /dev/null
fi

# macOS x64
if [ ! -f "$RESOURCES_DIR/ffmpeg-darwin-amd64" ]; then
    echo "Downloading FFmpeg for macOS x64..."
    curl -L -o "$RESOURCES_DIR/ffmpeg-darwin-amd64.zip" "https://evermeet.cx/ffmpeg/getrelease/intel/zip"
    cd "$RESOURCES_DIR"
    unzip ffmpeg-darwin-amd64.zip
    mv ffmpeg ffmpeg-darwin-amd64
    rm ffmpeg-darwin-amd64.zip
    cd - > /dev/null
fi

# Windows x64
if [ ! -f "$RESOURCES_DIR/ffmpeg-windows-amd64.exe" ]; then
    echo "Downloading FFmpeg for Windows x64..."
    curl -L -o "$RESOURCES_DIR/ffmpeg-windows.zip" "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip"
    cd "$RESOURCES_DIR"
    unzip -j ffmpeg-windows.zip "*/bin/ffmpeg.exe"
    mv ffmpeg.exe ffmpeg-windows-amd64.exe
    rm ffmpeg-windows.zip
    cd - > /dev/null
fi

# Linux x64
if [ ! -f "$RESOURCES_DIR/ffmpeg-linux-amd64" ]; then
    echo "Downloading FFmpeg for Linux x64..."
    curl -L -o "$RESOURCES_DIR/ffmpeg-linux.tar.xz" "https://johnvansickle.com/ffmpeg/builds/ffmpeg-git-amd64-static.tar.xz"
    cd "$RESOURCES_DIR"
    tar -xJf ffmpeg-linux.tar.xz --strip-components=1 --wildcards "*/ffmpeg"
    mv ffmpeg ffmpeg-linux-amd64
    rm ffmpeg-linux.tar.xz
    cd - > /dev/null
fi

echo "All FFmpeg binaries downloaded successfully!"
