package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"chandao-mini/backend/core/errors"
	"chandao-mini/backend/core/service"
)

// UserHandler 用户处理器
// 只负责HTTP请求/响应处理，业务逻辑由Service层处理
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetUsers 获取用户列表（支持分页）
// @Summary 获取用户列表
// @Description 获取禅道系统中的用户列表，支持分页
// @Tags 用户
// @Accept json
// @Produce json
// @Param page query int false "页码，默认1"
// @Param limit query int false "每页数量，默认20"
// @Success 200 {object} errors.Response{data=vo.PaginatedVO{list=[]vo.UserVO}}
// @Failure 500 {object} errors.Response
// @Router /api/v1/users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	// 获取分页参数
	page := 1
	limit := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// 调用Service层处理业务逻辑
	result, err := h.userService.GetUsers(page, limit)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	// 返回成功响应
	errors.Success(c, result)
}

// GetUsersAll 获取所有用户列表
// @Summary 获取所有用户列表
// @Description 获取禅道系统中的所有用户列表，缓存24小时
// @Tags 用户
// @Accept json
// @Produce json
// @Success 200 {object} errors.Response{data=map[string]interface{}}
// @Failure 500 {object} errors.Response
// @Router /api/v1/users/all [get]
func (h *UserHandler) GetUsersAll(c *gin.Context) {
	// 调用Service层处理业务逻辑
	users, err := h.userService.GetUsersAll()
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	// 返回成功响应
	errors.Success(c, map[string]interface{}{
		"users": users,
		"total": len(users),
	})
}

// GetCurrentUser 获取当前登录用户信息
// @Summary 获取当前登录用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户
// @Accept json
// @Produce json
// @Success 200 {object} errors.Response{data=map[string]interface{}}
// @Failure 500 {object} errors.Response
// @Router /api/v1/users/current [get]
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	// 调用Service层处理业务逻辑
	user, err := h.userService.GetCurrentUser()
	if err != nil {
		// 处理验证错误
		if _, ok := err.(*service.ValidationError); ok {
			errors.InternalError(c, err.Error())
			return
		}
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	// 返回成功响应
	errors.Success(c, map[string]interface{}{
		"user": user,
	})
}
