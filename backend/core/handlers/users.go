package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yi-nology/zentao-mini/backend/core/errors"
	"github.com/yi-nology/zentao-mini/backend/core/service"
)

type UserHandler struct {
	userService *service.UserService
}

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
// @Param pageSize query int false "每页数量，默认20"
// @Success 200 {object} errors.Response{data=vo.PaginatedVO{list=[]vo.UserVO}}
// @Failure 500 {object} errors.Response
// @Router /api/v1/users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	result, err := h.userService.GetUsers(page, pageSize)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

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
	users, err := h.userService.GetUsersAll()
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

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
	user, err := h.userService.GetCurrentUser()
	if err != nil {
		if _, ok := err.(*service.ValidationError); ok {
			errors.InternalError(c, err.Error())
			return
		}
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	errors.Success(c, map[string]interface{}{
		"user": user,
	})
}
