package handlers

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/yi-nology/zentao-mini/backend/core/dto"
	"github.com/yi-nology/zentao-mini/backend/core/errors"
	"github.com/yi-nology/zentao-mini/backend/core/service"
	"github.com/yi-nology/zentao-mini/backend/core/vo"
)

// 本文件为 Phase2b 新增的 7 个实体（cases/plans/programs/releases/testtasks/tickets/feedbacks）
// 提供 Handler。每个 Handler 一个 Get 方法，统一用 ExtendedQueryDTO 绑参。

// ---- 接口 ----

type CaseServicer interface {
	GetCasesByProduct(productID, page, pageSize int) (*vo.PaginatedVO, error)
}
type PlanServicer interface {
	GetPlans(productID, page, pageSize int) (*vo.PaginatedVO, error)
}
type ProgramServicer interface {
	GetPrograms(page, pageSize int) (*vo.PaginatedVO, error)
}
type ReleaseServicer interface {
	GetReleasesByProduct(productID, page, pageSize int) (*vo.PaginatedVO, error)
}
type TestTaskServicer interface {
	GetTestTasksByProduct(productID, page, pageSize int) (*vo.PaginatedVO, error)
}
type TicketServicer interface {
	GetTickets(browseType string, productID, page, pageSize int) (*vo.PaginatedVO, error)
}
type FeedbackServicer interface {
	GetFeedbacks(page, pageSize int) (*vo.PaginatedVO, error)
}

// ---- CaseHandler ----

type CaseHandler struct{ caseService CaseServicer }

func NewCaseHandler(s CaseServicer) *CaseHandler { return &CaseHandler{caseService: s} }

func (h *CaseHandler) GetCases(ctx context.Context, c *app.RequestContext) {
	var q dto.ExtendedQueryDTO
	if err := c.BindAndValidate(&q); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}
	q.Validate()
	result, err := h.caseService.GetCasesByProduct(q.ProductID, q.Page, q.PageSize)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}
	errors.Success(c, result)
}

// ---- PlanHandler ----

type PlanHandler struct{ planService PlanServicer }

func NewPlanHandler(s PlanServicer) *PlanHandler { return &PlanHandler{planService: s} }

func (h *PlanHandler) GetPlans(ctx context.Context, c *app.RequestContext) {
	var q dto.ExtendedQueryDTO
	if err := c.BindAndValidate(&q); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}
	q.Validate()
	result, err := h.planService.GetPlans(q.ProductID, q.Page, q.PageSize)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}
	errors.Success(c, result)
}

// ---- ProgramHandler（复用名字会冲突，用 ProjectProgramHandler）----

type ProgramHandler struct{ programService ProgramServicer }

func NewProgramHandler(s ProgramServicer) *ProgramHandler { return &ProgramHandler{programService: s} }

func (h *ProgramHandler) GetPrograms(ctx context.Context, c *app.RequestContext) {
	var q dto.ExtendedQueryDTO
	if err := c.BindAndValidate(&q); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}
	q.Validate()
	result, err := h.programService.GetPrograms(q.Page, q.PageSize)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}
	errors.Success(c, result)
}

// ---- ReleaseHandler ----

type ReleaseHandler struct{ releaseService ReleaseServicer }

func NewReleaseHandler(s ReleaseServicer) *ReleaseHandler { return &ReleaseHandler{releaseService: s} }

func (h *ReleaseHandler) GetReleases(ctx context.Context, c *app.RequestContext) {
	var q dto.ExtendedQueryDTO
	if err := c.BindAndValidate(&q); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}
	q.Validate()
	result, err := h.releaseService.GetReleasesByProduct(q.ProductID, q.Page, q.PageSize)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}
	errors.Success(c, result)
}

// ---- TestTaskHandler ----

type TestTaskHandler struct{ testTaskService TestTaskServicer }

func NewTestTaskHandler(s TestTaskServicer) *TestTaskHandler { return &TestTaskHandler{testTaskService: s} }

func (h *TestTaskHandler) GetTestTasks(ctx context.Context, c *app.RequestContext) {
	var q dto.ExtendedQueryDTO
	if err := c.BindAndValidate(&q); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}
	q.Validate()
	result, err := h.testTaskService.GetTestTasksByProduct(q.ProductID, q.Page, q.PageSize)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}
	errors.Success(c, result)
}

// ---- TicketHandler ----

type TicketHandler struct{ ticketService TicketServicer }

func NewTicketHandler(s TicketServicer) *TicketHandler { return &TicketHandler{ticketService: s} }

func (h *TicketHandler) GetTickets(ctx context.Context, c *app.RequestContext) {
	var q dto.ExtendedQueryDTO
	if err := c.BindAndValidate(&q); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}
	q.Validate()
	browseType := q.BrowseType
	if browseType == "" {
		browseType = "all"
	}
	result, err := h.ticketService.GetTickets(browseType, q.ProductID, q.Page, q.PageSize)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}
	errors.Success(c, result)
}

// ---- FeedbackHandler ----

type FeedbackHandler struct{ feedbackService FeedbackServicer }

func NewFeedbackHandler(s FeedbackServicer) *FeedbackHandler { return &FeedbackHandler{feedbackService: s} }

func (h *FeedbackHandler) GetFeedbacks(ctx context.Context, c *app.RequestContext) {
	var q dto.ExtendedQueryDTO
	if err := c.BindAndValidate(&q); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}
	q.Validate()
	result, err := h.feedbackService.GetFeedbacks(q.Page, q.PageSize)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}
	errors.Success(c, result)
}

// 编译期断言：确保 service 具体类型实现对应接口。
var (
	_ CaseServicer     = (*service.CaseService)(nil)
	_ PlanServicer     = (*service.PlanService)(nil)
	_ ProgramServicer  = (*service.ProgramService)(nil)
	_ ReleaseServicer  = (*service.ReleaseService)(nil)
	_ TestTaskServicer = (*service.TestTaskService)(nil)
	_ TicketServicer   = (*service.TicketService)(nil)
	_ FeedbackServicer = (*service.FeedbackService)(nil)
)
