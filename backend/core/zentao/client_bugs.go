package zentao

import (
	"context"
	"strconv"
	"time"

	"github.com/yi-nology/common/biz/zentao"
)

// GetBugs 获取产品的 Bug 列表
func (c *Client) GetBugs(productID int, page, pageSize int) ([]zentao.Bug, error) {
	cacheKey := DefaultKeyBuilder.Build("zentao:bugs", strconv.Itoa(productID), strconv.Itoa(page), strconv.Itoa(pageSize))

	result, err := GlobalCache.GetOrLoadWithLock(cacheKey, func() (interface{}, error) {
		var response *zentao.BugListResponse
		err := c.withTokenRetry("GetBugs", func(client *zentao.Client) error {
			var err error
			response, err = client.GetBugs(productID, page, pageSize)
			return err
		})
		if err != nil {
			return nil, err
		}
		return response.Bugs, nil
	}, 2*time.Minute)

	if err != nil {
		return nil, err
	}
	return result.([]zentao.Bug), nil
}

// GetBugsByProject 根据项目 ID 过滤 Bug 列表
func (c *Client) GetBugsByProject(productID, projectID int, page, pageSize int) ([]zentao.Bug, error) {
	var response *zentao.BugListResponse
	err := c.withTokenRetry("GetBugsByProject", func(client *zentao.Client) error {
		var err error
		response, err = client.GetBugsByProject(productID, projectID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Bugs, nil
}

// GetBugsByStatus 根据状态过滤 Bug 列表
func (c *Client) GetBugsByStatus(productID int, status string, page, pageSize int) ([]zentao.Bug, error) {
	var response *zentao.BugListResponse
	err := c.withTokenRetry("GetBugsByStatus", func(client *zentao.Client) error {
		var err error
		response, err = client.GetBugsByStatus(productID, status, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Bugs, nil
}

// SearchBugs 搜索 Bug（支持多条件过滤）
func (c *Client) SearchBugs(params zentao.BugSearchParams) ([]zentao.Bug, error) {
	var response *zentao.BugListResponse
	err := c.withTokenRetry("SearchBugs", func(client *zentao.Client) error {
		var err error
		response, err = client.SearchBugs(params)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Bugs, nil
}

// GetBug 获取 Bug 详情
func (c *Client) GetBug(bugID int) (*zentao.Bug, error) {
	var result *zentao.Bug
	err := c.withTokenRetry("GetBug", func(client *zentao.Client) error {
		var err error
		result, err = client.GetBug(bugID)
		return err
	})
	return result, err
}

// GetAllBugs 获取产品全部 Bug（自动翻页）
func (c *Client) GetAllBugs(productID int) ([]zentao.Bug, error) {
	var all []zentao.Bug
	page := 1
	for {
		bugs, err := c.GetBugs(productID, page, 100)
		if err != nil {
			return all, err
		}
		all = append(all, bugs...)
		if len(bugs) < 100 {
			break
		}
		page++
	}
	return all, nil
}

// GetAllBugsByProject 获取项目全部 Bug
func (c *Client) GetAllBugsByProject(projectID int) ([]zentao.Bug, error) {
	return c.GetAllBugsByProjectWithProduct(0, projectID)
}

// GetAllBugsByProjectWithProduct 获取项目全部 Bug（指定产品ID）
func (c *Client) GetAllBugsByProjectWithProduct(productID int, projectID int) ([]zentao.Bug, error) {
	var allBugs []zentao.Bug
	page := 1
	for {
		bugs, err := c.GetBugs(productID, page, 500)
		if err != nil {
			return allBugs, err
		}
		allBugs = append(allBugs, bugs...)
		if len(bugs) < 500 {
			break
		}
		page++
	}

	if projectID <= 0 {
		return allBugs, nil
	}
	filtered := make([]zentao.Bug, 0, len(allBugs))
	for _, bug := range allBugs {
		if bug.Project == projectID {
			filtered = append(filtered, bug)
		}
	}
	return filtered, nil
}

// GetBugsContext 获取 Bug 列表（支持 context 取消）
func (c *Client) GetBugsContext(ctx context.Context, productID int, page, pageSize int) ([]zentao.Bug, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var response *zentao.BugListResponse
	err := c.withTokenRetryContext(ctx, "GetBugs", func(client *zentao.Client) error {
		var err error
		response, err = client.GetBugs(productID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Bugs, nil
}

// GetAllBugsContext 获取产品全部 Bug（支持 context 取消，自动翻页）
func (c *Client) GetAllBugsContext(ctx context.Context, productID int) ([]zentao.Bug, error) {
	var all []zentao.Bug
	page := 1
	for {
		select {
		case <-ctx.Done():
			return all, ctx.Err()
		default:
		}
		bugs, err := c.GetBugsContext(ctx, productID, page, 100)
		if err != nil {
			return all, err
		}
		all = append(all, bugs...)
		if len(bugs) < 100 {
			break
		}
		page++
	}
	return all, nil
}

// GetAllBugsByProjectContext 获取项目全部 Bug（支持 context 取消）
func (c *Client) GetAllBugsByProjectContext(ctx context.Context, projectID int) ([]zentao.Bug, error) {
	var all []zentao.Bug
	page := 1
	for {
		select {
		case <-ctx.Done():
			return all, ctx.Err()
		default:
		}
		var response *zentao.BugListResponse
		err := c.withTokenRetryContext(ctx, "GetBugsByProject", func(client *zentao.Client) error {
			var err error
			response, err = client.GetBugsByProject(0, projectID, page, 100)
			return err
		})
		if err != nil {
			return all, err
		}
		all = append(all, response.Bugs...)
		if len(response.Bugs) < 100 {
			break
		}
		page++
	}
	return all, nil
}
