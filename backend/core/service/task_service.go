package service

import (
	"fmt"
	"time"

	"github.com/yi-nology/common/biz/zentao"
	"go.uber.org/zap"

	"github.com/yi-nology/zentao-mini/backend/core/dto"
	"github.com/yi-nology/zentao-mini/backend/core/logger"
	"github.com/yi-nology/zentao-mini/backend/core/utils"
	"github.com/yi-nology/zentao-mini/backend/core/vo"
	myzentao "github.com/yi-nology/zentao-mini/backend/core/zentao"
)

const (
	tasksCacheDuration = 60 * time.Minute
)

// TaskService 任务业务逻辑服务
// 负责处理任务相关的业务逻辑
type TaskService struct {
	client *myzentao.Client
}

// NewTaskService 创建任务服务
func NewTaskService(client *myzentao.Client) *TaskService {
	return &TaskService{client: client}
}

// GetTasks 获取任务列表
// 业务逻辑：
// 1. 如果指定执行ID，则查询该执行的任务
// 2. 如果未指定执行ID，则查询所有执行的任务
// 3. 应用筛选条件（指派人、状态、时间范围）
// 4. 分页处理
func (s *TaskService) GetTasks(query *dto.TaskQueryDTO) (*vo.PaginatedVO, error) {
	var allTasks []zentao.Task
	var err error

	if query.ExecutionID != 0 {
		cacheKey := fmt.Sprintf("tasks:execution:%d", query.ExecutionID)
		allTasks, err = s.getTasksWithCache(cacheKey, func() ([]zentao.Task, error) {
			return s.client.GetTasks(query.ExecutionID, 1, 10000)
		})
		if err != nil {
			return nil, err
		}
	} else {
		cacheKey := "tasks:all:product:" + fmt.Sprint(query.ProductID)
		allTasks, err = s.getTasksWithCache(cacheKey, func() ([]zentao.Task, error) {
			return s.fetchAllTasks(query.ProductID)
		})
		if err != nil {
			return nil, err
		}
	}

	// 使用链式过滤器进行筛选
	chainFilter := utils.NewChainFilter(allTasks)

	// 按指派人筛选
	if query.AssignedTo != "" {
		chainFilter = chainFilter.Filter(func(item zentao.Task) bool {
			return func() string {
				if ref, ok := item.AssignedTo.(zentao.UserRef); ok {
					return ref.Account
				}
				return ""
			}() == query.AssignedTo
		})
	}

	// 按状态筛选
	if query.Status != "" {
		chainFilter = chainFilter.Filter(func(item zentao.Task) bool {
			return item.Status == query.Status
		})
	}

	// 按日期范围筛选
	if query.StartDate != "" || query.EndDate != "" {
		chainFilter = chainFilter.FilterByDate(query.StartDate, query.EndDate, func(item zentao.Task) string {
			return item.OpenedDate
		})
	}

	// 获取总数
	total := chainFilter.Count()

	// 执行分页
	pagedTasks := chainFilter.Paginate(query.Page, query.PageSize).Result()

	list := s.convertToVO(pagedTasks)

	return &vo.PaginatedVO{
		List:     list,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

// convertToVO 将zentao.Task转换为vo.TaskVO
func (s *TaskService) convertToVO(tasks []zentao.Task) []vo.TaskVO {
	if len(tasks) == 0 {
		return []vo.TaskVO{}
	}

	result := make([]vo.TaskVO, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, vo.TaskVO{
			ID:           task.ID,
			Project:      task.Project,
			Execution:    task.Execution,
			Name:         task.Name,
			Type:         task.Type,
			Pri:          task.Pri,
			Status:       task.Status,
			AssignedTo:   task.AssignedTo,
			EstStarted:   task.EstStarted,
			Deadline:     task.Deadline,
			Estimate:     task.Estimate,
			Consumed:     task.Consumed,
			Left:         task.Left,
			Desc:         task.Desc,
			OpenedBy:     task.OpenedBy,
			OpenedDate:   task.OpenedDate,
			FinishedBy:   task.FinishedBy,
			FinishedDate: task.FinishedDate,
			ClosedBy:     task.ClosedBy,
			ClosedDate:   task.ClosedDate,
			StatusName:   task.StatusName,
		})
	}
	return result
}

// getTasksWithCache 带缓存获取任务
func (s *TaskService) getTasksWithCache(cacheKey string, loadFunc func() ([]zentao.Task, error)) ([]zentao.Task, error) {
	if cached, ok := myzentao.GlobalCache.Get(cacheKey); ok {
		if tasks, ok := cached.([]zentao.Task); ok {
			return tasks, nil
		}
	}

	tasks, err := loadFunc()
	if err != nil {
		return nil, err
	}

	myzentao.GlobalCache.Set(cacheKey, tasks, tasksCacheDuration)
	return tasks, nil
}

// fetchAllTasks 获取所有执行的任务
func (s *TaskService) fetchAllTasks(productId int) ([]zentao.Task, error) {
	var allTasks []zentao.Task

	projects, err := s.client.GetProjectsByProduct(productId, 1, 2000)
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		executions, err := s.client.GetExecutions(project.ID, 1, 1000)
		if err != nil {
			logger.Error("Failed to fetch executions for project",
				zap.Int("projectID", project.ID),
				zap.Error(err))
			continue
		}

		for _, execution := range executions {
			tasks, err := s.client.GetTasks(execution.ID, 1, 10000)
			if err != nil {
				logger.Error("Failed to fetch tasks for execution",
					zap.Int("executionID", execution.ID),
					zap.Error(err))
				continue
			}
			allTasks = append(allTasks, tasks...)
		}
	}

	return allTasks, nil
}
