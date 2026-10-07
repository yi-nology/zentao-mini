package service

import (
	"strings"

	"github.com/yi-nology/common/biz/zentao"

	"github.com/yi-nology/zentao-mini/backend/core/dto"
	"github.com/yi-nology/zentao-mini/backend/core/utils"
	"github.com/yi-nology/zentao-mini/backend/core/vo"
	myzentao "github.com/yi-nology/zentao-mini/backend/core/zentao"
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
	stories, err := s.fetchStories(query)
	if err != nil {
		return nil, err
	}

	chainFilter := filterStoryChain(stories, query)

	// 获取总数
	total := chainFilter.Count()

	// 执行分页
	pagedStories := chainFilter.Paginate(query.Page, query.PageSize).Result()

	list := s.convertToVO(pagedStories)

	return &vo.PaginatedVO{
		List:     list,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

// GetStoryStatusCounts 返回应用筛选条件后（不分页）的各状态需求数量，供统计卡使用
func (s *StoryService) GetStoryStatusCounts(query *dto.StoryQueryDTO) (map[string]int, error) {
	counts := map[string]int{"draft": 0, "active": 0, "reviewing": 0, "changed": 0, "closed": 0}

	stories, err := s.fetchStories(query)
	if err != nil {
		return nil, err
	}

	for _, story := range filterStoryChain(stories, query).Result() {
		if _, ok := counts[story.Status]; ok {
			counts[story.Status]++
		}
	}
	return counts, nil
}

// fetchStories 按优先级 executionID > projectID > productID 拉取需求全量
func (s *StoryService) fetchStories(query *dto.StoryQueryDTO) ([]zentao.Story, error) {
	if query.ExecutionID != 0 {
		return s.client.GetStoriesByExecution(query.ExecutionID, 1, 1000)
	}
	if query.ProjectID != 0 {
		return s.client.GetStoriesByProject(query.ProjectID, 1, 1000)
	}
	if query.ProductID != 0 {
		return s.client.GetStoriesByProduct(query.ProductID, 1, 1000)
	}
	return nil, &ValidationError{Message: "请提供产品ID、项目ID或执行ID"}
}

// filterStoryChain 对需求列表应用查询条件（不含分页），返回链式过滤器
func filterStoryChain(stories []zentao.Story, query *dto.StoryQueryDTO) *utils.ChainFilter[zentao.Story] {
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

	// 按状态筛选
	if query.Status != "" {
		chainFilter = chainFilter.Filter(func(item zentao.Story) bool {
			return strings.EqualFold(item.Status, query.Status)
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

	return chainFilter
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
