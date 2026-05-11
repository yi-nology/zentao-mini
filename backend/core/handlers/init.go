package handlers

import (
	"bufio"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"chandao-mini/backend/core/errors"
	"chandao-mini/backend/core/initialization"
	"chandao-mini/backend/core/logger"
	"chandao-mini/backend/core/zentao"
)

type InitHandler struct {
	initService  *initialization.InitService
	zentaoClient *zentao.Client
}

func NewInitHandler(initService *initialization.InitService, zentaoClient *zentao.Client) *InitHandler {
	return &InitHandler{
		initService:  initService,
		zentaoClient: zentaoClient,
	}
}

func (h *InitHandler) UploadConfig(c *gin.Context) {
	file, err := c.FormFile("configFile")
	if err != nil {
		errors.BadRequest(c, "请选择要上传的文件")
		return
	}

	fileData := make([]byte, file.Size)
	f, err := file.Open()
	if err != nil {
		errors.InternalError(c, "打开上传文件失败")
		return
	}
	defer f.Close()
	fileDataReader := bufio.NewReader(f)
	_, err = fileDataReader.Read(fileData)
	if err != nil {
		errors.InternalError(c, "读取上传文件失败")
		return
	}

	authConfig, err := h.initService.LoadEncryptedConfig(fileData)
	if err != nil {
		errors.InternalError(c, "加载认证配置失败")
		return
	}

	err = h.initService.StoreAuthConfig(fileData)
	if err != nil {
		errors.InternalError(c, "存储认证配置失败")
		return
	}

	h.zentaoClient.UpdateConfig(authConfig.Domain, authConfig.Username, authConfig.Password)

	errors.SuccessWithMessage(c, "配置已保存，正在后台连接禅道...", nil)
}

func (h *InitHandler) GetInitStatus(c *gin.Context) {
	status := h.initService.GetInitStatus()

	logger.Info("初始化状态检查",
		zap.Bool("isFirstStart", status.IsFirstStart),
		zap.Bool("hasConfig", status.HasConfig),
		zap.String("message", status.Message),
	)

	errors.Success(c, gin.H{
		"isFirstStart": status.IsFirstStart,
		"hasConfig":    status.HasConfig,
		"message":      status.Message,
	})
}

func (h *InitHandler) GetAccountInfo(c *gin.Context) {
	domain := h.zentaoClient.GetServer()
	account := h.zentaoClient.GetAccount()
	connected := h.zentaoClient.IsConnected()

	errors.Success(c, gin.H{
		"domain":    domain,
		"account":   account,
		"connected": connected,
	})
}
