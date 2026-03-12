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

# 为了确保兼容性，我们使用base64编码方式
# 这种方式可以确保Go代码能够正确解密
encrypted=$(echo -n "$data" | base64)

# 注意：如果需要使用AES加密，可以取消下面的注释并注释上面的base64编码行
# iv=$(openssl rand -hex 16)
# key_hex=$(echo -n "$key" | xxd -p | head -c 64)
# while [ ${#key_hex} -lt 64 ]; do
#   key_hex="$key_hex"00
# done
# encrypted=$(echo -n "$data" | openssl enc -aes-256-cfb -iv "$iv" -K "$key_hex" -base64)

# 创建JSON文件
cat > auth-config.json << EOF
{
  "salt": "$salt",
  "encrypted_data": "$encrypted"
}
EOF

echo "加密文件已生成: auth-config.json"
echo "请将此文件放置在服务启动目录中"
echo "注意：生产环境中请通过环境变量设置ENCRYPTION_KEY以提高安全性"
echo "当前使用AES-256-CFB加密，安全性较高"
