package handlers

import (
	"context"
	stderrors "errors"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/yi-nology/zentao-mini/backend/core/dto"
	"github.com/yi-nology/zentao-mini/backend/core/errors"
	"github.com/yi-nology/zentao-mini/backend/core/service"
)

type BugHandler struct {
	bugService BugServicer
}

func NewBugHandler(bugService BugServicer) *BugHandler {
	return &BugHandler{bugService: bugService}
}

func (h *BugHandler) GetBugs(ctx context.Context, c *app.RequestContext) {
	var query dto.BugQueryDTO
	if err := c.BindAndValidate(&query); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}

	if err := query.Validate(); err != nil {
		errors.Error(c, err)
		return
	}

	result, err := h.bugService.GetBugs(&query)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	errors.Success(c, result)
}

// AddBugComment POST /api/bugs/:id/comments 给 Bug 添加备注（需管理员登录）
func (h *BugHandler) AddBugComment(ctx context.Context, c *app.RequestContext) {
	bugID, err := service.ParseID(string(c.Param("id")))
	if err != nil {
		errors.BadRequest(c, err.Error())
		return
	}

	var req struct {
		Comment string `json:"comment"`
	}
	if err := c.BindJSON(&req); err != nil {
		errors.BadRequest(c, "请求格式错误")
		return
	}

	bug, err := h.bugService.AddBugComment(bugID, req.Comment)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	errors.SuccessWithMessage(c, "评论成功", bug)
}

// TransitionBug POST /api/bugs/:id/transitions 执行 Bug 状态流转（需管理员登录）
func (h *BugHandler) TransitionBug(ctx context.Context, c *app.RequestContext) {
	bugID, err := service.ParseID(string(c.Param("id")))
	if err != nil {
		errors.BadRequest(c, err.Error())
		return
	}

	var input service.BugTransitionInput
	if err := c.BindJSON(&input); err != nil {
		errors.BadRequest(c, "请求格式错误")
		return
	}

	bug, err := h.bugService.TransitionBug(bugID, &input)
	if err != nil {
		var verr *service.ValidationError
		if stderrors.As(err, &verr) {
			errors.BadRequest(c, verr.Error())
			return
		}
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	errors.SuccessWithMessage(c, "状态流转成功", bug)
}
