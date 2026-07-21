package zentao

import (
	"context"
	"strconv"
	"time"

	"github.com/yi-nology/common/biz/zentao"
)

// GetStoriesByProduct 获取产品的需求列表
func (c *Client) GetStoriesByProduct(productID int, page, pageSize int) ([]zentao.Story, error) {
	if c.IsSessionMode() {
		return c.getStoriesByProductSession(context.Background(), productID, page, pageSize)
	}
	cacheKey := DefaultKeyBuilder.Build("zentao:stories", strconv.Itoa(productID), strconv.Itoa(page), strconv.Itoa(pageSize))

	result, err := GlobalCache.GetOrLoadWithLock(cacheKey, func() (interface{}, error) {
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
	}, 2*time.Minute)

	if err != nil {
		return nil, err
	}
	return result.([]zentao.Story), nil
}

// GetStoriesByProject 获取项目的需求列表
func (c *Client) GetStoriesByProject(projectID int, page, pageSize int) ([]zentao.Story, error) {
	if c.IsSessionMode() {
		return c.getStoriesByProjectSession(context.Background(), projectID, page, pageSize)
	}
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
	if c.IsSessionMode() {
		return c.getStoriesByExecutionSession(context.Background(), executionID, page, pageSize)
	}
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

// GetAllStories 获取产品全部需求（自动翻页）
func (c *Client) GetAllStories(productID int) ([]zentao.Story, error) {
	if c.IsSessionMode() {
		return c.getAllStoriesSession(context.Background(), productID)
	}
	var all []zentao.Story
	page := 1
	for {
		stories, err := c.GetStoriesByProduct(productID, page, 100)
		if err != nil {
			return all, err
		}
		all = append(all, stories...)
		if len(stories) < 100 {
			break
		}
		page++
	}
	return all, nil
}

// GetAllStoriesByProject 获取项目全部需求（自动翻页）
func (c *Client) GetAllStoriesByProject(projectID int) ([]zentao.Story, error) {
	if c.IsSessionMode() {
		return c.getAllStoriesByProjectSession(context.Background(), projectID)
	}
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
	if c.IsSessionMode() {
		return c.getStoriesByProductSession(ctx, productID, page, pageSize)
	}
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

// GetAllStoriesContext 获取产品全部需求（支持 context 取消，自动翻页）
func (c *Client) GetAllStoriesContext(ctx context.Context, productID int) ([]zentao.Story, error) {
	if c.IsSessionMode() {
		return c.getAllStoriesSession(ctx, productID)
	}
	var all []zentao.Story
	page := 1
	for {
		select {
		case <-ctx.Done():
			return all, ctx.Err()
		default:
		}
		stories, err := c.GetStoriesByProductContext(ctx, productID, page, 100)
		if err != nil {
			return all, err
		}
		all = append(all, stories...)
		if len(stories) < 100 {
			break
		}
		page++
	}
	return all, nil
}
