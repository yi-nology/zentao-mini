package zentao

import (
	"context"
	"strconv"
	"time"

	"github.com/yi-nology/common/biz/zentao"
)

// GetAllProjects 获取所有项目列表
func (c *Client) GetAllProjects(limit int) ([]zentao.Project, error) {
	var response *zentao.ProjectListResponse
	err := c.withTokenRetry("GetAllProjects", func(client *zentao.Client) error {
		var err error
		response, err = client.GetAllProjects(1, limit)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Projects, nil
}

// GetProjectsByProduct 获取产品关联的项目列表
func (c *Client) GetProjectsByProduct(productID int, page, pageSize int) ([]zentao.Project, error) {
	cacheKey := DefaultKeyBuilder.Build("zentao:projects", strconv.Itoa(productID), strconv.Itoa(page), strconv.Itoa(pageSize))

	result, err := GlobalCache.GetOrLoadWithLock(cacheKey, func() (interface{}, error) {
		var response *zentao.ProjectListResponse
		err := c.withTokenRetry("GetProjectsByProduct", func(client *zentao.Client) error {
			var err error
			response, err = client.GetProjectsByProduct(productID, page, pageSize)
			return err
		})
		if err != nil {
			return nil, err
		}
		return response.Projects, nil
	}, 3*time.Minute)

	if err != nil {
		return nil, err
	}
	return result.([]zentao.Project), nil
}

// GetProject 获取项目详情
func (c *Client) GetProject(projectID int) (*zentao.Project, error) {
	var result *zentao.Project
	err := c.withTokenRetry("GetProject", func(client *zentao.Client) error {
		var err error
		result, err = client.GetProject(projectID)
		return err
	})
	return result, err
}

// GetProjectContext 获取项目详情（支持 context 取消）
func (c *Client) GetProjectContext(ctx context.Context, projectID int) (*zentao.Project, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var result *zentao.Project
	err := c.withTokenRetryContext(ctx, "GetProject", func(client *zentao.Client) error {
		var err error
		result, err = client.GetProject(projectID)
		return err
	})
	return result, err
}

// GetProjectsByProductContext 获取产品关联的项目（支持 context 取消）
func (c *Client) GetProjectsByProductContext(ctx context.Context, productID int, page, pageSize int) ([]zentao.Project, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var response *zentao.ProjectListResponse
	err := c.withTokenRetryContext(ctx, "GetProjectsByProduct", func(client *zentao.Client) error {
		var err error
		response, err = client.GetProjectsByProduct(productID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Projects, nil
}
