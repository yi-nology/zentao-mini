package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"chandao-mini/backend/core/models"
	"chandao-mini/backend/core/zentao"
)

// UserHandler 用户处理器
type UserHandler struct {
	client *zentao.Client
}

// NewUserHandler 创建用户处理器
func NewUserHandler(client *zentao.Client) *UserHandler {
	return &UserHandler{client: client}
}

// GetUsers 获取用户列表（支持分页）
// @Summary 获取用户列表
// @Description 获取禅道系统中的用户列表，支持分页
// @Tags 用户
// @Accept json
// @Produce json
// @Param page query int false "页码，默认1"
// @Param limit query int false "每页数量，默认20"
// @Success 200 {object} models.Response{data=map[string]interface{}}
// @Failure 500 {object} models.Response
// @Router /api/users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	// 获取分页参数
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")

	// 转换为整数
	pageInt := 1
	limitInt := 20
	if _, err := fmt.Sscanf(page, "%d", &pageInt); err != nil {
		pageInt = 1
	}
	if _, err := fmt.Sscanf(limit, "%d", &limitInt); err != nil {
		limitInt = 20
	}

	// 调用客户端方法
	userList, err := h.client.GetUsers(pageInt, limitInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "获取用户列表失败: " + err.Error(),
		})
		return
	}

	// 转换为模型
	var users []models.User
	for _, u := range userList.Users {
		users = append(users, models.User{
			ID:       u.ID,
			Account:  u.Account,
			Realname: u.Realname,
		})
	}

	// 构造响应数据
	responseData := map[string]interface{}{
		"users": users,
		"page":  userList.Page,
		"total": userList.Total,
		"limit": userList.Limit,
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "success",
		Data:    responseData,
	})
}

// GetUsersAll 获取所有用户列表
// @Summary 获取所有用户列表
// @Description 获取禅道系统中的所有用户列表，缓存24小时
// @Tags 用户
// @Accept json
// @Produce json
// @Success 200 {object} models.Response{data=map[string]interface{}}
// @Failure 500 {object} models.Response
// @Router /api/users/all [get]
func (h *UserHandler) GetUsersAll(c *gin.Context) {
	// 调用客户端方法，获取所有用户
	users, err := h.client.GetUsersAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "获取用户列表失败: " + err.Error(),
		})
		return
	}

	// 转换为模型
	var userModels []models.User
	for _, u := range users {
		userModels = append(userModels, models.User{
			ID:       u.ID,
			Account:  u.Account,
			Realname: u.Realname,
		})
	}

	// 构造响应数据
	responseData := map[string]interface{}{
		"users": userModels,
		"total": len(userModels),
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "success",
		Data:    responseData,
	})
}

// GetCurrentUser 获取当前登录用户信息
// @Summary 获取当前登录用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户
// @Accept json
// @Produce json
// @Success 200 {object} models.Response{data=map[string]interface{}}
// @Failure 500 {object} models.Response
// @Router /api/users/current [get]
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	// 调用客户端方法，获取所有用户
	users, err := h.client.GetUsersAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "获取用户列表失败: " + err.Error(),
		})
		return
	}

	// 查找当前登录用户
	var currentUser *models.User
	for _, u := range users {
		if u.Account == h.client.GetAccount() {
			currentUser = &models.User{
				ID:       u.ID,
				Account:  u.Account,
				Realname: u.Realname,
			}
			break
		}
	}

	if currentUser == nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "未找到当前登录用户",
		})
		return
	}

	// 构造响应数据
	responseData := map[string]interface{}{
		"user": currentUser,
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "success",
		Data:    responseData,
	})
}
