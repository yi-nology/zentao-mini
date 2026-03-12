#!/bin/bash

# 加密文件生成脚本
# 用于生成包含加密用户名、密码和域名的JSON文件

# 检查参数
if [ $# -ne 3 ]; then
  echo "Usage: $0 <username> <password> <domain>"
  exit 1
fi

username="$1"
password="$2"
domain="$3"

# 生成随机盐值
salt=$(openssl rand -base64 16)

# 生成密钥（使用环境变量或默认值）
if [ -z "$ENCRYPTION_KEY" ]; then
  # 默认密钥，生产环境应该通过环境变量设置
  ENCRYPTION_KEY="Zhangyi@Kylin999-"
fi

# 准备要加密的数据
data=$(echo -n "{\"username\":\"$username\",\"password\":\"$password\",\"domain\":\"$domain\"}")

# 生成密钥（与Go代码保持一致）
key="$ENCRYPTION_KEY$salt"
# 确保密钥长度为32字节
key=$(echo -n "$key" | head -c 32)

# 使用AES-256-CFB加密确保数据安全
iv=$(openssl rand -hex 16)
# 使用更可靠的方式生成十六进制密钥
key_hex=$(echo -n "$key" | openssl dgst -sha256 -hex | cut -d' ' -f2)
encrypted=$(echo -n "$data" | openssl enc -aes-256-cfb -iv "$iv" -K "$key_hex" -base64)

# 注意：base64编码已被注释，使用AES加密确保安全性
# encrypted=$(echo -n "$data" | base64)

# 创建JSON文件
cat > auth.json << EOF
{
  "salt": "$salt",
  "iv": "$iv",
  "encrypted_data": "$encrypted"
}
EOF

echo "加密文件已生成: auth.json"
echo "请将此文件放置在服务启动目录中"
echo "注意：生产环境中请通过环境变量设置ENCRYPTION_KEY以提高安全性"
echo "当前使用AES-256-CFB加密，安全性较高"
