package zentao

import (
	"context"
	"strconv"
	"time"

	"github.com/yi-nology/common/biz/zentao"
)

// GetExecutions 获取执行列表
func (c *Client) GetExecutions(projectID int, page, pageSize int) ([]zentao.Execution, error) {
	cacheKey := DefaultKeyBuilder.Build("zentao:executions", strconv.Itoa(projectID), strconv.Itoa(page), strconv.Itoa(pageSize))

	result, err := GlobalCache.GetOrLoadWithLock(cacheKey, func() (interface{}, error) {
		var response *zentao.ExecutionListResponse
		err := c.withTokenRetry("GetExecutions", func(client *zentao.Client) error {
			var err error
			response, err = client.GetExecutions(projectID, page, pageSize)
			return err
		})
		if err != nil {
			return nil, err
		}
		return response.Executions, nil
	}, 3*time.Minute)

	if err != nil {
		return nil, err
	}
	return result.([]zentao.Execution), nil
}

// GetExecutionsContext 获取执行列表（支持 context 取消）
func (c *Client) GetExecutionsContext(ctx context.Context, projectID int, page, pageSize int) ([]zentao.Execution, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var response *zentao.ExecutionListResponse
	err := c.withTokenRetryContext(ctx, "GetExecutions", func(client *zentao.Client) error {
		var err error
		response, err = client.GetExecutions(projectID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Executions, nil
}

// ExecutionContext 执行上下文（包含执行信息和关联的项目）
type ExecutionContext struct {
	Exec     zentao.Execution
	ProjName string
	ExecName string
}

// GetExecutionsByProduct 按产品获取所有执行
func (c *Client) GetExecutionsByProduct(productID int) ([]ExecutionContext, error) {
	var result []ExecutionContext
	projects, err := c.GetProjectsByProduct(productID, 1, 200)
	if err != nil {
		return nil, err
	}
	for _, p := range projects {
		executions, err := c.GetExecutions(p.ID, 1, 200)
		if err != nil {
			continue
		}
		for _, e := range executions {
			result = append(result, ExecutionContext{
				Exec:     e,
				ProjName: p.Name,
				ExecName: e.Name,
			})
		}
	}
	return result, nil
}

// GetExecutionsByProductContext 按产品获取所有执行（支持 context 取消）
func (c *Client) GetExecutionsByProductContext(ctx context.Context, productID int) ([]ExecutionContext, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var result []ExecutionContext
	projects, err := c.GetProjectsByProductContext(ctx, productID, 1, 200)
	if err != nil {
		return nil, err
	}
	for _, p := range projects {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}
		executions, err := c.GetExecutionsContext(ctx, p.ID, 1, 200)
		if err != nil {
			continue
		}
		for _, e := range executions {
			result = append(result, ExecutionContext{
				Exec:     e,
				ProjName: p.Name,
				ExecName: e.Name,
			})
		}
	}
	return result, nil
}
