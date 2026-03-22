package handlers

import (
	"net/http"

	"chandao-mini/backend/core/errors"
	"chandao-mini/backend/core/initialization"
	"chandao-mini/backend/core/zentao"

	"github.com/gin-gonic/gin"
)

// InitHandler 初始化处理器
type InitHandler struct {
	initService *initialization.InitService
	zentaoClient *zentao.Client
}

// NewInitHandler 创建初始化处理器
func NewInitHandler(initService *initialization.InitService, zentaoClient *zentao.Client) *InitHandler {
	return &InitHandler{
		initService: initService,
		zentaoClient: zentaoClient,
	}
}

// UploadConfig 上传配置文件
// @Summary 上传禅道认证配置文件
// @Description 接收加密的auth.json文件并存储
// @Tags init
// @Accept multipart/form-data
// @Param file formData file true "配置文件"
// @Success 200 {object} errors.Response
// @Failure 400 {object} errors.Response
// @Failure 500 {object} errors.Response
// @Router /api/init/upload [post]
func (h *InitHandler) UploadConfig(c *gin.Context) {
	file, err := c.FormFile("configFile")
	if err != nil {
		errors.BadRequest(c, "请上传配置文件")
		return
	}

	// 读取文件内容
	fileData, err := file.Open()
	if err != nil {
		errors.BadRequest(c, "无法读取文件")
		return
	}
	defer fileData.Close()

	// 读取全部数据
	content := make([]byte, file.Size)
	if _, err := fileData.Read(content); err != nil {
		errors.InternalError(c, "读取文件失败")
		return
	}

	// 存储配置
	if err := h.initService.StoreAuthConfig(content); err != nil {
		errors.InternalError(c, "存储配置失败: "+err.Error())
		return
	}

	errors.Success(c, gin.H{
		"message": "配置上传成功",
	})
}

// GetInitStatus 获取初始化状态
// @Summary 获取系统初始化状态
// @Description 检查系统是否已完成初始化配置
// @Tags init
// @Success 200 {object} errors.Response{data=initialization.InitStatus}
// @Failure 500 {object} errors.Response
// @Router /api/init/status [get]
func (h *InitHandler) GetInitStatus(c *gin.Context) {
	status := h.initService.GetInitStatus()
	errors.Success(c, status)
}

// GetHealth 健康检查
// @Summary 健康检查
// @Description 检查服务是否正常运行
// @Tags health
// @Success 200 {object} errors.Response
// @Router /health [get]
func (h *InitHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
