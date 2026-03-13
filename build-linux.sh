#!/bin/bash

# 交叉编译 Wails 应用为 x86_64 Linux 可执行文件

echo "开始交叉编译 Wails 应用为 x86_64 Linux 版本..."

# 构建前端
echo "构建前端..."
cd frontend && npm run build:wails
if [ $? -ne 0 ]; then
    echo "错误: 前端构建失败"
    exit 1
fi
cd ..

# 交叉编译为 Linux x86_64
echo "交叉编译为 Linux x86_64..."

# 创建输出目录
mkdir -p build/linux-amd64

# 使用 go build 直接交叉编译，添加 Wails 构建标签
GOOS=linux GOARCH=amd64 go build -tags wails -o build/linux-amd64/chandao-mini .
if [ $? -ne 0 ]; then
    echo "错误: 交叉编译失败"
    exit 1
fi

# 复制环境变量文件
cp .env.wails build/linux-amd64/.env

# 验证文件是否生成
if [ -f "build/linux-amd64/chandao-mini" ]; then
    echo "交叉编译完成!"
    echo "可执行文件位置: build/linux-amd64/chandao-mini"
    echo "文件大小: $(ls -lh build/linux-amd64/chandao-mini | awk '{print $5}')"
    echo "环境变量文件已复制: build/linux-amd64/.env"
else
    echo "错误: 可执行文件未生成"
    exit 1
fi
