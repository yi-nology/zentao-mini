#!/bin/bash

# 交叉编译 Wails 应用为 Windows x64 可执行文件

echo "开始交叉编译 Wails 应用为 Windows x64 版本..."

# 构建前端
echo "构建前端..."
cd frontend && npm run build:wails
if [ $? -ne 0 ]; then
    echo "错误: 前端构建失败"
    exit 1
fi
cd ..

# 交叉编译为 Windows x64
echo "交叉编译为 Windows x64..."

# 创建输出目录
mkdir -p build/windows-x64

# 使用 go build 直接交叉编译，添加 Wails 构建标签
GOOS=windows GOARCH=amd64 go build -tags wails -o build/windows-x64/chandao-mini.exe .
if [ $? -ne 0 ]; then
    echo "错误: 交叉编译失败"
    exit 1
fi

# 复制环境变量文件
cp .env.wails build/windows-x64/.env

echo "交叉编译完成!"
echo "可执行文件位置: build/windows-x64/chandao-mini.exe"
echo "环境变量文件已复制: build/windows-x64/.env"
