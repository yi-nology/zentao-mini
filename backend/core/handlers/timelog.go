package handlers

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/yi-nology/zentao-mini/backend/core/dto"
	"github.com/yi-nology/zentao-mini/backend/core/errors"
)

type TimelogHandler struct {
	timelogService TimelogServicer
}

func NewTimelogHandler(timelogService TimelogServicer) *TimelogHandler {
	return &TimelogHandler{timelogService: timelogService}
}

func (h *TimelogHandler) GetTimelogAnalysis(ctx context.Context, c *app.RequestContext) {
	var query dto.TimelogQueryDTO
	if err := c.BindAndValidate(&query); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}

	if query.ProductID == "" {
		errors.BadRequest(c, "productId is required")
		return
	}

	if query.DateFrom == "" || query.DateTo == "" {
		errors.BadRequest(c, "dateFrom and dateTo are required")
		return
	}

	result, err := h.timelogService.GetTimelogAnalysis(&query)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	errors.Success(c, result)
}

func (h *TimelogHandler) GetTimelogDashboard(ctx context.Context, c *app.RequestContext) {
	var query dto.TimelogQueryDTO
	if err := c.BindAndValidate(&query); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}

	if query.ProductID == "" {
		errors.BadRequest(c, "productId is required")
		return
	}

	if query.DateFrom == "" || query.DateTo == "" {
		errors.BadRequest(c, "dateFrom and dateTo are required")
		return
	}

	result, err := h.timelogService.GetTimelogDashboard(&query)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	errors.Success(c, result)
}

func (h *TimelogHandler) GetTimelogEfforts(ctx context.Context, c *app.RequestContext) {
	var query dto.TimelogQueryDTO
	if err := c.BindAndValidate(&query); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}

	if query.ProductID == "" {
		errors.BadRequest(c, "productId is required")
		return
	}

	if query.DateFrom == "" || query.DateTo == "" {
		errors.BadRequest(c, "dateFrom and dateTo are required")
		return
	}

	result, err := h.timelogService.GetTimelogEfforts(&query)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	errors.Success(c, result)
}
