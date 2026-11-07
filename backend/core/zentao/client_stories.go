package zentao

import (
	"context"
	"strconv"
	"time"

	"github.com/oekit/pkg/pagination"
	"github.com/yi-nology/common/biz/zentao"
)

// GetStoriesByProduct 获取产品的需求列表
func (c *Client) GetStoriesByProduct(productID int, page, pageSize int) ([]zentao.Story, error) {
	cacheKey := DefaultKeyBuilder.Build("zentao:stories", strconv.Itoa(productID), strconv.Itoa(page), strconv.Itoa(pageSize))

	return withCachedList(cacheKey, 2*time.Minute, func() ([]zentao.Story, error) {
		var response *zentao.StoryListResponse
		err := c.withTokenRetry("GetStoriesByProduct", func(client *zentao.Client) error {
			var err error
			response, err = client.GetStoriesByProduct(productID, page, pageSize)
			return err
		})
		if err != nil {
			return nil, err
		}
		return response.Stories, nil
	})
}

// GetStoriesByProject 获取项目的需求列表
func (c *Client) GetStoriesByProject(projectID int, page, pageSize int) ([]zentao.Story, error) {
	var response *zentao.StoryListResponse
	err := c.withTokenRetry("GetStoriesByProject", func(client *zentao.Client) error {
		var err error
		response, err = client.GetStoriesByProject(projectID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Stories, nil
}

// GetStoriesByExecution 获取执行的需求列表
func (c *Client) GetStoriesByExecution(executionID int, page, pageSize int) ([]zentao.Story, error) {
	var response *zentao.StoryListResponse
	err := c.withTokenRetry("GetStoriesByExecution", func(client *zentao.Client) error {
		var err error
		response, err = client.GetStoriesByExecution(executionID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Stories, nil
}

// GetStory 获取需求详情
func (c *Client) GetStory(storyID int) (*zentao.Story, error) {
	var result *zentao.Story
	err := c.withTokenRetry("GetStory", func(client *zentao.Client) error {
		var err error
		result, err = client.GetStory(storyID)
		return err
	})
	return result, err
}

// GetAllStories 获取产品全部需求（自动翻页，pagination.Walk 短页即末页）
func (c *Client) GetAllStories(productID int) ([]zentao.Story, error) {
	return pagination.Walk(context.Background(), 100, func(_ context.Context, page int) ([]zentao.Story, error) {
		return c.GetStoriesByProduct(productID, page, 100)
	})
}

// GetAllStoriesByProject 获取项目全部需求（自动翻页）
func (c *Client) GetAllStoriesByProject(projectID int) ([]zentao.Story, error) {
	var all []zentao.Story
	executions, err := c.GetExecutions(projectID, 1, 1000)
	if err != nil {
		return nil, err
	}
	for _, execution := range executions {
		stories, err := c.GetStoriesByExecution(execution.ID, 1, 10000)
		if err != nil {
			continue
		}
		all = append(all, stories...)
	}
	return all, nil
}

// GetStoriesByProductContext 获取产品的需求列表（支持 context 取消）
func (c *Client) GetStoriesByProductContext(ctx context.Context, productID int, page, pageSize int) ([]zentao.Story, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var response *zentao.StoryListResponse
	err := c.withTokenRetryContext(ctx, "GetStoriesByProduct", func(client *zentao.Client) error {
		var err error
		response, err = client.GetStoriesByProduct(productID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Stories, nil
}

// GetAllStoriesContext 获取产品全部需求（支持 context 取消，pagination.Walk 短页即末页）
func (c *Client) GetAllStoriesContext(ctx context.Context, productID int) ([]zentao.Story, error) {
	return pagination.Walk(ctx, 100, func(ctx context.Context, page int) ([]zentao.Story, error) {
		return c.GetStoriesByProductContext(ctx, productID, page, 100)
	})
}
