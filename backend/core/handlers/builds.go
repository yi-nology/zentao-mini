package handlers

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/yi-nology/zentao-mini/backend/core/dto"
	"github.com/yi-nology/zentao-mini/backend/core/errors"
)

type BuildHandler struct {
	buildService BuildServicer
}

func NewBuildHandler(buildService BuildServicer) *BuildHandler {
	return &BuildHandler{buildService: buildService}
}

func (h *BuildHandler) GetBuilds(ctx context.Context, c *app.RequestContext) {
	var query dto.BuildQueryDTO
	if err := c.BindAndValidate(&query); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}

	if err := query.Validate(); err != nil {
		errors.Error(c, err)
		return
	}

	result, err := h.buildService.GetBuilds(&query)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	errors.Success(c, result)
}
