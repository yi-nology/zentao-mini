package service

import (
	"strings"

	"github.com/yi-nology/common/biz/zentao"

	"chandao-mini/backend/core/dto"
	"chandao-mini/backend/core/utils"
	"chandao-mini/backend/core/vo"
	myzentao "chandao-mini/backend/core/zentao"
)

// StoryService 需求业务逻辑服务
// 负责处理需求相关的业务逻辑
type StoryService struct {
	client *myzentao.Client
}

// NewStoryService 创建需求服务
func NewStoryService(client *myzentao.Client) *StoryService {
	return &StoryService{client: client}
}

// GetStories 获取需求列表
// 业务逻辑：
// 1. 根据执行ID、项目ID或产品ID查询需求（优先级：executionID > projectID > productID）
// 2. 应用筛选条件（指派人、时间范围）
// 3. 分页处理
func (s *StoryService) GetStories(query *dto.StoryQueryDTO) (*vo.PaginatedVO, error) {
	var stories []zentao.Story
	var err error

	// 优先级: executionID > projectID > productID
	if query.ExecutionID != 0 {
		stories, err = s.client.GetStoriesByExecution(query.ExecutionID, 500)
	} else if query.ProjectID != 0 {
		stories, err = s.client.GetStoriesByProject(query.ProjectID, 500)
	} else if query.ProductID != 0 {
		stories, err = s.client.GetStoriesByProduct(query.ProductID, 500)
	} else {
		return nil, &ValidationError{Message: "请提供产品ID、项目ID或执行ID"}
	}

	if err != nil {
		return nil, err
	}

	// 使用链式过滤器进行筛选
	chainFilter := utils.NewChainFilter(stories)

	// 按指派人筛选
	if query.AssignedTo != "" {
		chainFilter = chainFilter.Filter(func(item zentao.Story) bool {
			assignedStr := ""
			if item.AssignedTo != nil {
				switch v := item.AssignedTo.(type) {
				case string:
					assignedStr = v
				case map[string]interface{}:
					if account, ok := v["account"].(string); ok {
						assignedStr = account
					}
				}
			}
			return strings.EqualFold(assignedStr, query.AssignedTo)
		})
	}

	// 按时间范围或具体日期筛选
	if query.StartDate != "" || query.EndDate != "" || query.SpecificDate != "" {
		chainFilter = chainFilter.Filter(func(item zentao.Story) bool {
			filtered := utils.FilterByDateRangeOrSpecific(
				[]zentao.Story{item},
				query.StartDate,
				query.EndDate,
				query.SpecificDate,
				func(s zentao.Story) string { return s.OpenedDate },
			)
			return len(filtered) > 0
		})
	}

	// 获取总数
	total := chainFilter.Count()

	// 执行分页
	pagedStories := chainFilter.Paginate(query.Page, query.Limit).Result()

	// 转换为VO
	list := s.convertToVO(pagedStories)

	// 返回分页结果
	return &vo.PaginatedVO{
		List:  list,
		Total: total,
		Page:  query.Page,
		Limit: query.Limit,
	}, nil
}

// convertToVO 将zentao.Story转换为vo.StoryVO
func (s *StoryService) convertToVO(stories []zentao.Story) []vo.StoryVO {
	if len(stories) == 0 {
		return []vo.StoryVO{}
	}

	result := make([]vo.StoryVO, 0, len(stories))
	for _, story := range stories {
		result = append(result, vo.StoryVO{
			ID:           story.ID,
			Product:      story.Product,
			Module:       story.Module,
			Plan:         story.Plan,
			Source:       story.Source,
			Title:        story.Title,
			Spec:         story.Spec,
			Verify:       story.Verify,
			Type:         story.Type,
			Status:       story.Status,
			Stage:        story.Stage,
			Pri:          story.Pri,
			Estimate:     story.Estimate,
			Version:      story.Version,
			OpenedBy:     story.OpenedBy,
			OpenedDate:   story.OpenedDate,
			AssignedTo:   story.AssignedTo,
			AssignedDate: story.AssignedDate,
			ClosedBy:     story.ClosedBy,
			ClosedDate:   story.ClosedDate,
			ClosedReason: story.ClosedReason,
		})
	}
	return result
}

// ValidationError 验证错误
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
