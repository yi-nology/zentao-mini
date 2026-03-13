package service

import (
	"github.com/yi-nology/common/biz/zentao"

	"chandao-mini/backend/core/dto"
	"chandao-mini/backend/core/utils"
	"chandao-mini/backend/core/vo"
	myzentao "chandao-mini/backend/core/zentao"
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
// 1. 根据执行ID查询任务
// 2. 应用筛选条件（指派人、状态、时间范围）
// 3. 分页处理
func (s *TaskService) GetTasks(query *dto.TaskQueryDTO) (*vo.PaginatedVO, error) {
	// 获取任务列表
	tasks, err := s.client.GetTasks(query.ExecutionID, 500)
	if err != nil {
		return nil, err
	}

	// 使用链式过滤器进行筛选
	chainFilter := utils.NewChainFilter(tasks)

	// 按指派人筛选
	if query.AssignedTo != "" {
		chainFilter = chainFilter.Filter(func(item zentao.Task) bool {
			return item.AssignedTo.Account == query.AssignedTo
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
	pagedTasks := chainFilter.Paginate(query.Page, query.Limit).Result()

	// 转换为VO
	list := s.convertToVO(pagedTasks)

	// 返回分页结果
	return &vo.PaginatedVO{
		List:  list,
		Total: total,
		Page:  query.Page,
		Limit: query.Limit,
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
			AssignedTo:   vo.UserRefVO(task.AssignedTo),
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
