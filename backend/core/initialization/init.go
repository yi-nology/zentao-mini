package initialization

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// AuthConfig 认证配置结构
type AuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Domain   string `json:"domain"`
}

// EncryptedAuthConfig 加密的认证配置结构
type EncryptedAuthConfig struct {
	Salt          string `json:"salt"`
	Iv            string `json:"iv"`
	EncryptedData string `json:"encrypted_data"`
}

// InitService 初始化服务
type InitService struct {
	authConfigPath string
	dbPath         string
	encryptionKey  string
}

// NewInitService 创建初始化服务实例
func NewInitService(authConfigPath, dbPath, encryptionKey string) *InitService {
	if authConfigPath == "" {
		authConfigPath = "./auth-config.json"
	}
	if dbPath == "" {
		// 使用用户主目录作为存储位置，确保在打包应用中也能正确访问
		homeDir, err := os.UserHomeDir()
		if err == nil {
			dbPath = filepath.Join(homeDir, ".zentao-mini", "auth.db")
		} else {
			// 如果无法获取用户主目录，使用当前目录
			dbPath = "./auth.db"
		}
	}
	if encryptionKey == "" {
		// 从环境变量读取加密密钥，如果未设置则使用默认值（仅用于开发环境）
		encryptionKey = os.Getenv("ZENTAO_ENCRYPTION_KEY")
		if encryptionKey == "" {
			// 生产环境必须设置环境变量，这里给出警告
			log.Println("WARNING: ZENTAO_ENCRYPTION_KEY environment variable is not set. Using default key for development only.")
			log.Println("WARNING: Please set ZENTAO_ENCRYPTION_KEY environment variable in production!")
			encryptionKey = "dev-default-key-change-in-production"
		}
	}

	return &InitService{
		authConfigPath: authConfigPath,
		dbPath:         dbPath,
		encryptionKey:  encryptionKey,
	}
}

// IsFirstStart 检测是否为首次启动
func (s *InitService) IsFirstStart() (bool, error) {
	// 检查数据库文件是否存在
	_, err := os.Stat(s.dbPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 数据库文件不存在，判定为首次启动
			return true, nil
		}
		return false, err
	}

	// 检查数据库文件是否为空
	fileInfo, err := os.Stat(s.dbPath)
	if err != nil {
		return false, err
	}

	if fileInfo.Size() == 0 {
		// 数据库文件为空，判定为首次启动
		return true, nil
	}

	// 数据库文件存在且不为空，判定为非首次启动
	return false, nil
}

// LoadEncryptedConfig 加载加密配置文件
func (s *InitService) LoadEncryptedConfig(fileData []byte) (*AuthConfig, error) {

	// 解析加密配置
	var encryptedConfig EncryptedAuthConfig
	if err := json.Unmarshal(fileData, &encryptedConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal encrypted config: %w", err)
	}

	// 解密数据
	authConfig, err := s.decrypt(encryptedConfig.EncryptedData, encryptedConfig.Salt, encryptedConfig.Iv)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt auth config: %w", err)
	}

	return authConfig, nil
}

// StoreAuthConfig 存储认证配置到数据库
func (s *InitService) StoreAuthConfig(fileData []byte) error {

	// 确保目录存在
	dbDir := filepath.Dir(s.dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create db directory: %w", err)
	}

	// 写入数据库文件（存储加密的JSON文件原文）
	if err := os.WriteFile(s.dbPath, fileData, 0600); err != nil {
		return fmt.Errorf("failed to write auth config to db: %w", err)
	}

	return nil
}

// LoadAuthConfig 从数据库加载认证配置
func (s *InitService) LoadAuthConfig() (*AuthConfig, []byte, error) {
	// 读取数据库文件（存储的是加密的JSON文件原文）
	fileData, err := os.ReadFile(s.dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read auth config from db: %w", err)
	}

	// 解析加密配置
	var encryptedConfig EncryptedAuthConfig
	if err := json.Unmarshal(fileData, &encryptedConfig); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal encrypted config: %w", err)
	}

	// 解密数据
	authConfig, err := s.decrypt(encryptedConfig.EncryptedData, encryptedConfig.Salt, encryptedConfig.Iv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt auth config: %w", err)
	}

	return authConfig, fileData, nil
}

// decrypt 解密数据
func (s *InitService) decrypt(encryptedData, salt, iv string) (*AuthConfig, error) {
	// 解码base64数据
	decodedData, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	// 尝试直接解析为AuthConfig（处理base64编码的情况）
	var config AuthConfig
	if err := json.Unmarshal(decodedData, &config); err == nil {
		// 直接解析成功，返回结果
		return &config, nil
	}

	// 尝试AES解密（处理真正加密的情况）
	// 检查IV是否提供
	if iv == "" {
		return nil, errors.New("IV is required for AES decryption")
	}

	// 解码IV
	ivBytes, err := hex.DecodeString(iv)
	if err != nil {
		return nil, fmt.Errorf("failed to decode IV: %w", err)
	}

	// 创建密钥（与shell脚本保持一致，使用SHA-256哈希）
	key := []byte(s.encryptionKey + salt)
	// 确保密钥长度为32字节（与shell脚本保持一致）
	if len(key) > 32 {
		key = key[:32]
	}
	keyHash := sha256.Sum256(key)
	keyBytes := keyHash[:]

	// 创建AES加密块
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}

	// 检查密文长度
	if len(decodedData) == 0 {
		return nil, errors.New("ciphertext is empty")
	}

	// 创建解密器
	stream := cipher.NewCFBDecrypter(block, ivBytes)

	// 解密数据
	plaintext := make([]byte, len(decodedData))
	stream.XORKeyStream(plaintext, decodedData)

	// 解析认证配置
	if err := json.Unmarshal(plaintext, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Encrypt 加密数据（用于测试）
func (s *InitService) Encrypt(config *AuthConfig, salt string) (string, error) {
	// 序列化认证配置
	configData, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	// 创建密钥
	key := []byte(s.encryptionKey + salt)
	if len(key) < 32 {
		// 填充密钥到32字节
		for len(key) < 32 {
			key = append(key, 0)
		}
	} else if len(key) > 32 {
		// 截断密钥到32字节
		key = key[:32]
	}

	// 创建AES加密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 创建IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// 创建加密器
	stream := cipher.NewCFBEncrypter(block, iv)

	// 加密数据
	ciphertext := make([]byte, len(configData))
	stream.XORKeyStream(ciphertext, configData)

	// 组合IV和密文
	ciphertext = append(iv, ciphertext...)

	// 编码为base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// LoadZentaoConfig 加载禅道配置
func LoadZentaoConfig(initService *InitService) (string, string, string) {
	// 检测是否为首次启动
	isFirstStart, err := initService.IsFirstStart()
	if err != nil {
		log.Printf("Warning: Failed to check first start status: %v", err)
		// 继续执行，使用环境变量
	}

	// 尝试从配置源加载配置
	var authConfig *AuthConfig
	var loadErr error
	var fileData []byte

	if isFirstStart {
		log.Println("First start detected, loading encrypted config...")
	} else {
		log.Println("Non-first start, loading config from db...")
		// 非首次启动，尝试从数据库加载配置
		authConfig, fileData, loadErr = initService.LoadAuthConfig()
		if loadErr != nil {
			log.Printf("Error loading auth config from db: %v", loadErr)
		}
	}

	// 如果配置加载成功，使用配置中的值
	if authConfig != nil {
		// 如果是首次启动，存储配置到数据库
		if isFirstStart {
			if err := initService.StoreAuthConfig(fileData); err != nil {
				log.Printf("Error storing auth config to db: %v", err)
			} else {
				log.Println("Auth config stored to db successfully")
			}
		}
		// 打印账号和域名，不打印密码
		log.Printf("Loaded config: Username=%s, Domain=%s", authConfig.Username, authConfig.Domain)
		return authConfig.Domain, authConfig.Username, authConfig.Password
	}

	// 配置加载失败，使用环境变量
	log.Println("Using environment variables for zentao config")
	zentaoServer := os.Getenv("ZENTAO_SERVER")
	zentaoAccount := os.Getenv("ZENTAO_ACCOUNT")
	zentaoPassword := os.Getenv("ZENTAO_PASSWORD")

	return zentaoServer, zentaoAccount, zentaoPassword
}
