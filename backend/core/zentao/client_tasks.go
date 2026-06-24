package zentao

import (
	"context"
	"strconv"
	"time"

	"github.com/yi-nology/common/biz/zentao"
)

// GetTasks 获取执行的任务列表
func (c *Client) GetTasks(executionID int, page, pageSize int) ([]zentao.Task, error) {
	cacheKey := DefaultKeyBuilder.Build("zentao:tasks", strconv.Itoa(executionID), strconv.Itoa(page), strconv.Itoa(pageSize))

	result, err := GlobalCache.GetOrLoadWithLock(cacheKey, func() (interface{}, error) {
		var response *zentao.TaskListResponse
		err := c.withTokenRetry("GetTasks", func(client *zentao.Client) error {
			var err error
			response, err = client.GetTasks(executionID, page, pageSize)
			return err
		})
		if err != nil {
			return nil, err
		}
		return response.Tasks, nil
	}, 2*time.Minute)

	if err != nil {
		return nil, err
	}
	return result.([]zentao.Task), nil
}

// GetTask 获取任务详情
func (c *Client) GetTask(taskID int) (*zentao.Task, error) {
	var result *zentao.Task
	err := c.withTokenRetry("GetTask", func(client *zentao.Client) error {
		var err error
		result, err = client.GetTask(taskID)
		return err
	})
	return result, err
}

// GetAllTasksByProject 获取项目全部任务（自动翻页）
func (c *Client) GetAllTasksByProject(projectID int) ([]zentao.Task, error) {
	var all []zentao.Task
	executions, err := c.GetExecutions(projectID, 1, 1000)
	if err != nil {
		return nil, err
	}
	for _, execution := range executions {
		tasks, err := c.GetTasks(execution.ID, 1, 10000)
		if err != nil {
			continue
		}
		all = append(all, tasks...)
	}
	return all, nil
}

// GetAllTasksByProduct 获取产品全部任务（自动翻页）
func (c *Client) GetAllTasksByProduct(productID int) ([]zentao.Task, error) {
	var all []zentao.Task
	projects, err := c.GetProjectsByProduct(productID, 1, 2000)
	if err != nil {
		return nil, err
	}
	for _, project := range projects {
		tasks, err := c.GetAllTasksByProject(project.ID)
		if err != nil {
			continue
		}
		all = append(all, tasks...)
	}
	return all, nil
}

// GetTasksContext 获取执行的任务列表（支持 context 取消）
func (c *Client) GetTasksContext(ctx context.Context, executionID int, page, pageSize int) ([]zentao.Task, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var response *zentao.TaskListResponse
	err := c.withTokenRetryContext(ctx, "GetTasks", func(client *zentao.Client) error {
		var err error
		response, err = client.GetTasks(executionID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Tasks, nil
}
