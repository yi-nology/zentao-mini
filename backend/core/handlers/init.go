package handlers

import (
	"bufio"
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"

	"github.com/yi-nology/zentao-mini/backend/core/errors"
	"github.com/yi-nology/zentao-mini/backend/core/initialization"
	"github.com/yi-nology/zentao-mini/backend/core/logger"
	"github.com/yi-nology/zentao-mini/backend/core/zentao"
)

type InitHandler struct {
	initService  *initialization.InitService
	zentaoClient *zentao.Client
}

// loginRequest 账号密码登录表单请求体。
type loginRequest struct {
	Domain   string `json:"domain"`
	Account  string `json:"account"`
	Password string `json:"password"`
	// Realm 认证域：非空（如 "kydc"）走会话模式，空走 Token 模式。
	Realm string `json:"realm"`
}

func NewInitHandler(initService *initialization.InitService, zentaoClient *zentao.Client) *InitHandler {
	return &InitHandler{
		initService:  initService,
		zentaoClient: zentaoClient,
	}
}

func (h *InitHandler) UploadConfig(ctx context.Context, c *app.RequestContext) {
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

// Login 账号密码登录端点（POST /api/init/login）。
// 与 UploadConfig（加密文件）并存的另一种初始化方式。
//
// 流程：
//  1. 校验 domain/account/password 非空
//  2. 调 zentaoClient.UpdateSessionConfig 或 UpdateConfig 触发实际登录验证
//  3. 登录成功 → 加密持久化 AuthConfig（含 realm）到 auth.db
//  4. 登录失败 → 返回 400 + 后端原始错误（不触发前端 401 重定向）
//
// 注意：先验证再持久化，避免把错误凭据写盘。
func (h *InitHandler) Login(ctx context.Context, c *app.RequestContext) {
	var req loginRequest
	if err := c.BindAndValidate(&req); err != nil {
		errors.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	if req.Domain == "" || req.Account == "" || req.Password == "" {
		errors.BadRequest(c, "domain、account、password 不能为空")
		return
	}

	// 先尝试验证凭据（热重载客户端）。失败立即返回，不落盘。
	if req.Realm != "" {
		if err := h.zentaoClient.UpdateSessionConfig(req.Domain, req.Account, req.Password, req.Realm); err != nil {
			logger.Warn("会话登录失败",
				zap.String("domain", req.Domain),
				zap.String("account", req.Account),
				zap.String("realm", req.Realm),
				zap.Error(err),
			)
			errors.BadRequest(c, "登录失败: "+err.Error())
			return
		}
	} else {
		h.zentaoClient.UpdateConfig(req.Domain, req.Account, req.Password)
	}

	// 验证通过，持久化。注意 Session 模式下 connected 已被 UpdateSessionConfig 置 true。
	authConfig := &initialization.AuthConfig{
		Username: req.Account,
		Password: req.Password,
		Domain:   req.Domain,
		Realm:    req.Realm,
	}
	if err := h.initService.StoreAuthConfigFromStruct(authConfig); err != nil {
		logger.Error("持久化登录配置失败", zap.Error(err))
		errors.InternalError(c, "保存登录配置失败")
		return
	}

	mode := "token"
	if req.Realm != "" {
		mode = "session"
	}
	errors.SuccessWithMessage(c, "登录成功", map[string]interface{}{
		"domain":    h.zentaoClient.GetServer(),
		"account":   h.zentaoClient.GetAccount(),
		"connected": h.zentaoClient.IsConnected(),
		"mode":      mode,
		"realm":     req.Realm,
	})
}

func (h *InitHandler) GetInitStatus(ctx context.Context, c *app.RequestContext) {
	status := h.initService.GetInitStatus()

	logger.Info("初始化状态检查",
		zap.Bool("isFirstStart", status.IsFirstStart),
		zap.Bool("hasConfig", status.HasConfig),
		zap.String("message", status.Message),
	)

	errors.Success(c, map[string]interface{}{
		"isFirstStart": status.IsFirstStart,
		"hasConfig":    status.HasConfig,
		"message":      status.Message,
	})
}

func (h *InitHandler) GetAccountInfo(ctx context.Context, c *app.RequestContext) {
	domain := h.zentaoClient.GetServer()
	account := h.zentaoClient.GetAccount()
	connected := h.zentaoClient.IsConnected()

	mode := "token"
	realm := ""
	if h.zentaoClient.IsSessionMode() {
		mode = "session"
		realm = h.zentaoClient.GetRealm()
	}

	errors.Success(c, map[string]interface{}{
		"domain":    domain,
		"account":   account,
		"connected": connected,
		"mode":      mode,
		"realm":     realm,
	})
}
