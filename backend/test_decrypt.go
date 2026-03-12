package main

import (
	"fmt"
	"os"

	"chandao-mini/backend/core/initialization"
)

func main() {
	// 读取auth.json文件
	data, err := os.ReadFile("../auth.json")
	if err != nil {
		fmt.Printf("Error reading auth.json: %v\n", err)
		return
	}

	// 创建初始化服务实例
	initService := initialization.NewInitService("", "", "")

	// 尝试解密
	authConfig, err := initService.LoadEncryptedConfig(data)
	if err != nil {
		fmt.Printf("Error decrypting config: %v\n", err)
		return
	}

	// 打印解密结果
	fmt.Println("Decrypted config:")
	fmt.Printf("Username: %s\n", authConfig.Username)
	fmt.Printf("Password: %s\n", authConfig.Password)
	fmt.Printf("Domain: %s\n", authConfig.Domain)

	fmt.Println("\nVerification successful!")
}
