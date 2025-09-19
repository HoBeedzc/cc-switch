#!/bin/bash

# Build script for cc-switch
set -e

echo "Building cc-switch for multiple platforms..."

# 创建bin目录
mkdir -p bin

# 构建配置
PLATFORMS=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
)

# 构建每个平台
for platform in "${PLATFORMS[@]}"; do
  IFS='/' read -r GOOS GOARCH <<< "$platform"
  
  echo "Building for $GOOS/$GOARCH..."
  
  # 设置输出文件名
  if [ "$GOOS" = "windows" ]; then
    output="bin/cc-switch-$GOOS-$GOARCH.exe"
  else
    output="bin/cc-switch-$GOOS-$GOARCH"
  fi
  
  # 构建
  GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "$output" .
  
  echo "✓ Built $output"
done

echo ""
echo "✓ All builds completed successfully!"
echo ""
echo "Binary files:"
ls -la bin/