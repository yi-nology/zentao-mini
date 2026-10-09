package service

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/yi-nology/common/biz/zentao"

	"github.com/yi-nology/zentao-mini/backend/core/dto"
	"github.com/yi-nology/zentao-mini/backend/core/utils"
	"github.com/yi-nology/zentao-mini/backend/core/vo"
	myzentao "github.com/yi-nology/zentao-mini/backend/core/zentao"
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
// 2. 应用筛选条件（状态、指派人、版本、时间范围）
// 3. 分页处理
func (s *BugService) GetBugs(query *dto.BugQueryDTO) (*vo.PaginatedVO, error) {
	var bugs []zentao.Bug
	var err error

	// 如果有产品ID，按产品查询
	if query.ProductID != 0 {
		// 指派人/版本/类型/状态过滤需要获取所有bug（含closed），在内存中过滤
		// 注意：closed 状态的 bug 不会被禅道默认接口返回，
		// 必须用 status=all 全量获取（含 closed）后在内存过滤
		if query.AssignedTo != "" || query.Version != "" || query.Type != "" || query.Status != "" {
			bugs, err = s.client.GetAllBugsIncludeClosed(query.ProductID)
		} else if query.ProjectID != 0 {
			// 如果只有项目ID，使用GetBugsByProject
			bugs, err = s.client.GetBugsByProject(query.ProductID, query.ProjectID, 1, 1000)
		} else {
			// 获取产品的所有Bug
			bugs, err = s.client.GetBugs(query.ProductID, 1, 1000)
		}

		if err != nil {
			return nil, err
		}
	} else {
		// 如果没有产品ID，返回空列表
		bugs = []zentao.Bug{}
	}

	chainFilter := filterBugChain(bugs, query)

	total := chainFilter.Count()
	paged := chainFilter.Paginate(query.Page, query.PageSize).Result()

	list := s.convertToVO(paged)
	if query.Lite {
		applyLiteSteps(list)
	}

	return &vo.PaginatedVO{
		List:     list,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

// severityToInt 禅道 Bug.severity 是 interface{}（JSON 数字/字符串双形态实弹在案：
// 上游不同接口返回形态不一），归一为 int；nil/非法值返回 0。
func severityToInt(v interface{}) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(t))
		if err != nil {
			return 0
		}
		return n
	}
	return 0
}

// liteSteps 轻量 Steps：剥 HTML 标签、折叠空白、超 300 字符截断并带指引标记。
// 列表面（MCP）不需要完整富文本——实弹 2026-10-08：单产品 76 条 active 全量
// steps 即 689KB，超出引擎上下文承载被截断，真正的重现步骤详情应走 get_bug。
func liteSteps(s string) string {
	stripped := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(s, " ")
	stripped = strings.Join(strings.Fields(stripped), " ")
	const maxLen = 300
	if len([]rune(stripped)) <= maxLen {
		return stripped
	}
	runes := []rune(stripped)
	return string(runes[:maxLen]) + "…［步骤已截断，完整重现步骤用 get_bug 按 ID 取］"
}

// applyLiteSteps 原地轻量化列表 VO 的 Steps 字段。
func applyLiteSteps(list []vo.BugVO) {
	for i := range list {
		list[i].Steps = liteSteps(list[i].Steps)
	}
}

// GetBugStatusCounts 返回应用筛选条件后（不分页）的各状态 Bug 数量，供统计卡使用。
// 统计基于含 closed 的全量数据，否则"已关闭"永远是 0。
func (s *BugService) GetBugStatusCounts(query *dto.BugQueryDTO) (map[string]int, error) {
	counts := map[string]int{"active": 0, "resolved": 0, "closed": 0}
	if query.ProductID == 0 {
		return counts, nil
	}

	bugs, err := s.client.GetAllBugsIncludeClosed(query.ProductID)
	if err != nil {
		return nil, err
	}

	for _, bug := range filterBugChain(bugs, query).Result() {
		if _, ok := counts[bug.Status]; ok {
			counts[bug.Status]++
		}
	}
	return counts, nil
}

// filterBugChain 对全量 bug 列表应用查询条件（不含分页），返回链式过滤器.
func filterBugChain(bugs []zentao.Bug, query *dto.BugQueryDTO) *utils.ChainFilter[zentao.Bug] {
	chainFilter := utils.NewChainFilter(bugs)

	// 按状态筛选
	if query.Status != "" {
		chainFilter = chainFilter.Filter(func(item zentao.Bug) bool {
			return item.Status == query.Status
		})
	}

	// 按指派人筛选
	if query.AssignedTo != "" {
		chainFilter = chainFilter.Filter(func(item zentao.Bug) bool {
			return item.AssignedTo.Account == query.AssignedTo
		})
	}

	// 按严重度筛选（1-5；severity 在模型里是 interface{}，经 severityToInt 归一）
	if query.Severity != 0 {
		chainFilter = chainFilter.Filter(func(item zentao.Bug) bool {
			return severityToInt(item.Severity) == query.Severity
		})
	}

	// 按版本筛选（openedBuild 包含指定版本名称）
	if query.Version != "" {
		chainFilter = chainFilter.Filter(func(item zentao.Bug) bool {
			for _, build := range item.OpenedBuild {
				if build == query.Version {
					return true
				}
			}
			return false
		})
	}

	// 按类型筛选
	if query.Type != "" {
		chainFilter = chainFilter.Filter(func(item zentao.Bug) bool {
			return item.Type == query.Type
		})
	}

	// 按时间范围或具体日期筛选（创建日期 openedDate）
	if query.StartDate != "" || query.EndDate != "" || query.SpecificDate != "" {
		chainFilter = chainFilter.Filter(func(item zentao.Bug) bool {
			filtered := utils.FilterByDateRangeOrSpecific(
				[]zentao.Bug{item},
				query.StartDate,
				query.EndDate,
				query.SpecificDate,
				func(b zentao.Bug) string { s, _ := b.OpenedDate.(string); return s },
			)
			return len(filtered) > 0
		})
	}

	// 按解决日期窗筛选（resolvedDate；扁鹊批次二百三十四：「今天解决了多少」的承接面，
	// 与 StartDate/EndDate 作用的 openedDate 相互独立可叠加——「今天新增且今天解决」）。
	if query.ResolvedStartDate != "" || query.ResolvedEndDate != "" {
		chainFilter = chainFilter.Filter(func(item zentao.Bug) bool {
			filtered := utils.FilterByDateRange(
				[]zentao.Bug{item},
				query.ResolvedStartDate,
				query.ResolvedEndDate,
				func(b zentao.Bug) string { s, _ := b.ResolvedDate.(string); return s },
			)
			return len(filtered) > 0
		})
	}

	// 按关闭日期窗筛选（closedDate；v1.6.1，215 实弹 sess-1009-zsyb6euc 专家拍板点：
	// 736 条已关闭 Bug 无 closedDate 过滤参数，「今日关闭数」只能全量翻页——本窗补齐
	// 第三日期域。注意与 GetBugs 的取数前提：closed 状态 Bug 须 status 过滤/全量含关闭
	// 接口才返回，纯缺省列表不含 closed）。
	if query.ClosedStartDate != "" || query.ClosedEndDate != "" {
		chainFilter = chainFilter.Filter(func(item zentao.Bug) bool {
			filtered := utils.FilterByDateRange(
				[]zentao.Bug{item},
				query.ClosedStartDate,
				query.ClosedEndDate,
				func(b zentao.Bug) string { s, _ := b.ClosedDate.(string); return s },
			)
			return len(filtered) > 0
		})
	}

	return chainFilter
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
