#!/bin/bash

# 交叉编译 Wails 应用为 x86_64 Linux 可执行文件
#
# 注意：本脚本使用 CGO_ENABLED=0 进行纯 Go 交叉编译
#   - SQLite 离线缓存使用 modernc.org/sqlite（纯 Go 驱动），无需 CGO
#   - 应用菜单/桌面通知通过 Wails v2 内置功能，无需额外系统库
#   - 不包含原生系统托盘（Linux 需 libayatana-appindicator + CGO）
#     用户通过 OS 任务栏图标 / 应用菜单恢复隐藏的窗口
#
# 本地编译（在 Linux 上启用 GTK 完整 GUI）请使用 `wails build`
# 那需要安装 libgtk-3-dev libwebkit2gtk-4.1-dev pkg-config 等

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
# CGO_ENABLED=0 保证纯 Go 编译，不依赖目标系统的 C 库
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags wails -o build/linux-amd64/zentao-mini .
if [ $? -ne 0 ]; then
    echo "错误: 交叉编译失败"
    exit 1
fi

# 复制环境变量文件
cp frontend/.env.wails build/linux-amd64/.env

# 验证文件是否生成
if [ -f "build/linux-amd64/zentao-mini" ]; then
    echo "交叉编译完成!"
    echo "可执行文件位置: build/linux-amd64/zentao-mini"
    echo "文件大小: $(ls -lh build/linux-amd64/zentao-mini | awk '{print $5}')"
    echo "环境变量文件已复制: build/linux-amd64/.env"
else
    echo "错误: 可执行文件未生成"
    exit 1
fi
