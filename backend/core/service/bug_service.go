package service

import (
	"github.com/yi-nology/common/biz/zentao"

	"chandao-mini/backend/core/dto"
	"chandao-mini/backend/core/utils"
	"chandao-mini/backend/core/vo"
	myzentao "chandao-mini/backend/core/zentao"
)

// BugService Bug业务逻辑服务
// 负责处理Bug相关的业务逻辑
type BugService struct {
	client *myzentao.Client
}

// NewBugService 创建Bug服务
func NewBugService(client *myzentao.Client) *BugService {
	return &BugService{client: client}
}

// GetBugs 获取Bug列表
// 业务逻辑：
// 1. 根据产品ID查询Bug
// 2. 应用筛选条件（状态、指派人、时间范围）
// 3. 分页处理
func (s *BugService) GetBugs(query *dto.BugQueryDTO) (*vo.PaginatedVO, error) {
	var bugs []zentao.Bug
	var err error

	// 如果有产品ID，按产品查询
	if query.ProductID != 0 {
		// 优先使用SearchBugs进行多条件筛选，减少内存消耗
		if query.AssignedTo != "" || query.Status != "" {
			params := zentao.BugSearchParams{
				ProductID:  query.ProductID,
				Status:     query.Status,
				AssignedTo: query.AssignedTo,
				Limit:      1000, // 一次获取足够多的数据用于筛选
				Page:       1,
			}
			bugs, err = s.client.SearchBugs(params)
		} else if query.ProjectID != 0 {
			// 如果只有项目ID，使用GetBugsByProject
			bugs, err = s.client.GetBugsByProject(query.ProductID, query.ProjectID, 1000)
		} else {
			// 获取产品的所有Bug
			bugs, err = s.client.GetBugs(query.ProductID, 1000)
		}

		if err != nil {
			return nil, err
		}
	} else {
		// 如果没有产品ID，返回空列表
		bugs = []zentao.Bug{}
	}

	// 使用链式过滤器进行筛选
	chainFilter := utils.NewChainFilter(bugs)

	// 按时间范围或具体日期筛选
	chainFilter = chainFilter.Filter(func(item zentao.Bug) bool {
		return utils.FilterByDateRangeOrSpecific(
			[]zentao.Bug{item},
			query.StartDate,
			query.EndDate,
			query.SpecificDate,
			func(b zentao.Bug) string { return b.OpenedDate },
		) != nil
	})

	// 获取总数
	total := chainFilter.Count()

	// 执行分页
	pagedBugs := chainFilter.Paginate(query.Page, query.Limit).Result()

	// 转换为VO
	list := s.convertToVO(pagedBugs)

	// 返回分页结果
	return &vo.PaginatedVO{
		List:  list,
		Total: total,
		Page:  query.Page,
		Limit: query.Limit,
	}, nil
}

// convertToVO 将zentao.Bug转换为vo.BugVO
func (s *BugService) convertToVO(bugs []zentao.Bug) []vo.BugVO {
	if len(bugs) == 0 {
		return []vo.BugVO{}
	}

	result := make([]vo.BugVO, 0, len(bugs))
	for _, bug := range bugs {
		result = append(result, vo.BugVO{
			ID:            bug.ID,
			Project:       bug.Project,
			Product:       bug.Product,
			Title:         bug.Title,
			Keywords:      bug.Keywords,
			Severity:      bug.Severity,
			Pri:           bug.Pri,
			Type:          bug.Type,
			OS:            bug.OS,
			Browser:       bug.Browser,
			Hardware:      bug.Hardware,
			Steps:         bug.Steps,
			Status:        bug.Status,
			SubStatus:     bug.SubStatus,
			Color:         bug.Color,
			Confirmed:     bug.Confirmed,
			PlanTime:      bug.PlanTime,
			OpenedBy:      vo.UserRefVO(bug.OpenedBy),
			OpenedDate:    bug.OpenedDate,
			OpenedBuild:   bug.OpenedBuild,
			AssignedTo:    vo.UserRefVO(bug.AssignedTo),
			AssignedDate:  bug.AssignedDate,
			Deadline:      bug.Deadline,
			ResolvedBy:    bug.ResolvedBy,
			Resolution:    bug.Resolution,
			ResolvedBuild: bug.ResolvedBuild,
			ResolvedDate:  bug.ResolvedDate,
			ClosedBy:      bug.ClosedBy,
			ClosedDate:    bug.ClosedDate,
			StatusName:    bug.StatusName,
			LifeCycle:     bug.LifeCycle,
		})
	}
	return result
}
