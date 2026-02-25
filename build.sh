#!/bin/bash
set -e

OUTPUT_DIR="versions"
APP_NAME="gfood-printer"

mkdir -p "$OUTPUT_DIR"

echo "🔨 Gerando executáveis em ./$OUTPUT_DIR/ ..."

# macOS — Apple Silicon (M1/M2/M3)
echo "  → macOS ARM64 (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -o "$OUTPUT_DIR/${APP_NAME}-mac-arm"

# macOS — Intel
echo "  → macOS AMD64 (Intel)..."
GOOS=darwin GOARCH=amd64 go build -o "$OUTPUT_DIR/${APP_NAME}-mac-intel"

# Linux — 64-bit
echo "  → Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -o "$OUTPUT_DIR/${APP_NAME}-linux"

# Linux — ARM64 (Raspberry Pi, servidores ARM)
echo "  → Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -o "$OUTPUT_DIR/${APP_NAME}-linux-arm64"

# Windows — 64-bit
echo "  → Windows AMD64 (64-bit)..."
GOOS=windows GOARCH=amd64 go build -o "$OUTPUT_DIR/${APP_NAME}-x64.exe"

# Windows — 32-bit
echo "  → Windows 386 (32-bit)..."
GOOS=windows GOARCH=386 go build -o "$OUTPUT_DIR/${APP_NAME}-x86.exe"

echo ""
echo "✅ Concluído! Executáveis disponíveis em ./$OUTPUT_DIR/:"
ls -lh "$OUTPUT_DIR/"
