#!/bin/bash

# 交叉编译 Wails 应用为 arm64 Linux 可执行文件

echo "开始交叉编译 Wails 应用为 arm64 Linux 版本..."

# 构建前端
echo "构建前端..."
cd frontend && npm run build
if [ $? -ne 0 ]; then
    echo "错误: 前端构建失败"
    exit 1
fi
cd ..

# 交叉编译为 Linux arm64
echo "交叉编译为 Linux arm64..."

# 创建输出目录
mkdir -p build/linux-arm64

# 使用 go build 直接交叉编译，添加 Wails 构建标签
GOOS=linux GOARCH=arm64 go build -tags wails -o build/linux-arm64/chandao-mini .
if [ $? -ne 0 ]; then
    echo "错误: 交叉编译失败"
    exit 1
fi

echo "交叉编译完成!"
echo "可执行文件位置: build/linux-arm64/chandao-mini"
