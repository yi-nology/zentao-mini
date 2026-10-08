package service

import (
	"strings"
	"testing"

	"github.com/yi-nology/common/biz/zentao"

	"github.com/yi-nology/zentao-mini/backend/core/dto"
	"github.com/yi-nology/zentao-mini/backend/core/utils"
	"github.com/yi-nology/zentao-mini/backend/core/vo"
)

func createMockBugs() []zentao.Bug {
	return []zentao.Bug{
		{
			ID:         1,
			Project:    10,
			Product:    100,
			Title:      "Bug 1",
			Status:     "active",
			OpenedDate: "2024-01-15 10:00:00",
			AssignedTo: zentao.UserRef{Account: "user1", Realname: "User 1"},
			OpenedBy:   zentao.UserRef{Account: "user2", Realname: "User 2"},
		},
		{
			ID:         2,
			Project:    10,
			Product:    100,
			Title:      "Bug 2",
			Status:     "resolved",
			OpenedDate: "2024-01-16 11:00:00",
			AssignedTo: zentao.UserRef{Account: "user1", Realname: "User 1"},
			OpenedBy:   zentao.UserRef{Account: "user3", Realname: "User 3"},
		},
		{
			ID:         3,
			Project:    11,
			Product:    100,
			Title:      "Bug 3",
			Status:     "active",
			OpenedDate: "2024-01-17 12:00:00",
			AssignedTo: zentao.UserRef{Account: "user2", Realname: "User 2"},
			OpenedBy:   zentao.UserRef{Account: "user1", Realname: "User 1"},
		},
	}
}

// TestBugService_ConvertToVO tests the VO conversion logic
func TestBugService_ConvertToVO(t *testing.T) {
	service := &BugService{client: nil}

	bugs := createMockBugs()
	vos := service.convertToVO(bugs)

	if len(vos) != 3 {
		t.Errorf("expected 3 VOs, got %d", len(vos))
	}

	if vos[0].ID != 1 {
		t.Errorf("expected first VO ID=1, got %d", vos[0].ID)
	}
	if vos[0].Title != "Bug 1" {
		t.Errorf("expected first VO Title='Bug 1', got '%s'", vos[0].Title)
	}
	if vos[0].Status != "active" {
		t.Errorf("expected first VO Status='active', got '%s'", vos[0].Status)
	}
}

// TestBugService_ConvertToVO_Empty tests empty slice conversion
func TestBugService_ConvertToVO_Empty(t *testing.T) {
	service := &BugService{client: nil}

	vos := service.convertToVO([]zentao.Bug{})

	if len(vos) != 0 {
		t.Errorf("expected 0 VOs, got %d", len(vos))
	}
}

// TestBugService_DateFilterLogic tests the date filter logic used in GetBugs
func TestBugService_DateFilterLogic(t *testing.T) {
	bugs := createMockBugs()

	tests := []struct {
		name          string
		startDate     string
		endDate       string
		specificDate  string
		expectedCount int
	}{
		{
			name:          "no date filter",
			expectedCount: 3,
		},
		{
			name:          "filter by date range",
			startDate:     "2024-01-15",
			endDate:       "2024-01-16",
			expectedCount: 2,
		},
		{
			name:          "filter by specific date",
			specificDate:  "2024-01-17",
			expectedCount: 1,
		},
		{
			name:          "filter with no matches",
			startDate:     "2025-01-01",
			endDate:       "2025-01-31",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chainFilter := utils.NewChainFilter(bugs)

			if tt.startDate != "" || tt.endDate != "" || tt.specificDate != "" {
				chainFilter = chainFilter.Filter(func(item zentao.Bug) bool {
					filtered := utils.FilterByDateRangeOrSpecific(
						[]zentao.Bug{item},
						tt.startDate,
						tt.endDate,
						tt.specificDate,
						func(b zentao.Bug) string { s, _ := b.OpenedDate.(string); return s },
					)
					return len(filtered) > 0
				})
			}

			if chainFilter.Count() != tt.expectedCount {
				t.Errorf("expected %d bugs, got %d", tt.expectedCount, chainFilter.Count())
			}
		})
	}
}

// TestBugService_Pagination tests the pagination logic
func TestBugService_Pagination(t *testing.T) {
	bugs := createMockBugs()

	tests := []struct {
		name          string
		page          int
		pageSize      int
		expectedLen   int
		expectedTotal int
	}{
		{
			name:          "page 1 size 2",
			page:          1,
			pageSize:      2,
			expectedLen:   2,
			expectedTotal: 3,
		},
		{
			name:          "page 2 size 2",
			page:          2,
			pageSize:      2,
			expectedLen:   1,
			expectedTotal: 3,
		},
		{
			name:          "page 1 size 10",
			page:          1,
			pageSize:      10,
			expectedLen:   3,
			expectedTotal: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chainFilter := utils.NewChainFilter(bugs)
			total := chainFilter.Count()
			paged := chainFilter.Paginate(tt.page, tt.pageSize).Result()

			if total != tt.expectedTotal {
				t.Errorf("expected total=%d, got %d", tt.expectedTotal, total)
			}
			if len(paged) != tt.expectedLen {
				t.Errorf("expected len=%d, got %d", tt.expectedLen, len(paged))
			}
		})
	}
}

// TestBugService_FilterBugs tests the in-memory filtering pipeline used by GetBugs.
func TestBugService_FilterBugs(t *testing.T) {
	bugs := createMockBugs()
	// 追加一个 closed 状态的 bug，验证 assignedTo 过滤也能覆盖 closed.
	const closedStatus = "closed"
	bugs = append(bugs, zentao.Bug{
		ID:         4,
		Project:    10,
		Product:    100,
		Title:      "Bug 4",
		Status:     closedStatus,
		OpenedDate: "2024-01-18 09:00:00",
		AssignedTo: zentao.UserRef{Account: "user1", Realname: "User 1"},
		OpenedBy:   zentao.UserRef{Account: "user2", Realname: "User 2"},
	})

	tests := []struct {
		name          string
		query         dto.BugQueryDTO
		expectedIDs   []int
		expectedTotal int
	}{
		{
			name:          "assignedTo only",
			query:         dto.BugQueryDTO{AssignedTo: "user1"},
			expectedIDs:   []int{1, 2, 4},
			expectedTotal: 3,
		},
		{
			name:          "assignedTo + status combined",
			query:         dto.BugQueryDTO{AssignedTo: "user1", Status: "active"},
			expectedIDs:   []int{1},
			expectedTotal: 1,
		},
		{
			name:          "assignedTo with no matches",
			query:         dto.BugQueryDTO{AssignedTo: "nosuchuser"},
			expectedIDs:   []int{},
			expectedTotal: 0,
		},
		{
			name:          "status only",
			query:         dto.BugQueryDTO{Status: "active"},
			expectedIDs:   []int{1, 3},
			expectedTotal: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.query
			q.Page = 1
			q.PageSize = 20
			paged := filterBugChain(bugs, &q).Paginate(q.Page, q.PageSize).Result()

			if len(paged) != len(tt.expectedIDs) {
				t.Fatalf("expected %d bugs, got %d", len(tt.expectedIDs), len(paged))
			}
			for i, want := range tt.expectedIDs {
				if paged[i].ID != want {
					t.Errorf("expected paged[%d].ID=%d, got %d", i, want, paged[i].ID)
				}
			}
		})
	}
}

// TestBugService_StatusCounts_FilterChain 统计基于过滤链结果按状态计数
func TestBugService_StatusCounts_FilterChain(t *testing.T) {
	bugs := createMockBugs()
	bugs = append(bugs, zentao.Bug{ID: 4, Status: "closed", AssignedTo: zentao.UserRef{Account: "user1"}})

	counts := map[string]int{"active": 0, "resolved": 0, "closed": 0}
	for _, bug := range filterBugChain(bugs, &dto.BugQueryDTO{AssignedTo: "user1"}).Result() {
		if _, ok := counts[bug.Status]; ok {
			counts[bug.Status]++
		}
	}
	if counts["active"] != 1 || counts["resolved"] != 1 || counts["closed"] != 1 {
		t.Errorf("expected active=1 resolved=1 closed=1, got %v", counts)
	}
}

// TestBugService_StatusCounts_NoProduct 未选产品时统计应全为 0（与空列表口径一致）
func TestBugService_StatusCounts_NoProduct(t *testing.T) {
	service := &BugService{client: nil}
	counts, err := service.GetBugStatusCounts(&dto.BugQueryDTO{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if counts["active"] != 0 || counts["resolved"] != 0 || counts["closed"] != 0 {
		t.Errorf("expected all zero counts, got %v", counts)
	}
}

// TestBugService_VOTypes tests that VO types are correctly mapped
func TestBugService_VOTypes(t *testing.T) {
	service := &BugService{client: nil}

	bug := zentao.Bug{
		ID:         42,
		Project:    10,
		Product:    100,
		Title:      "Test Bug",
		Status:     "active",
		OpenedDate: "2024-01-15 10:00:00",
		AssignedTo: zentao.UserRef{Account: "user1", Realname: "User 1"},
		OpenedBy:   zentao.UserRef{Account: "user2", Realname: "User 2"},
	}

	vos := service.convertToVO([]zentao.Bug{bug})
	if len(vos) != 1 {
		t.Fatalf("expected 1 VO, got %d", len(vos))
	}

	voItem := vos[0]
	if voItem.ID != 42 {
		t.Errorf("expected VO ID=42, got %d", voItem.ID)
	}
	if voItem.Title != "Test Bug" {
		t.Errorf("expected VO Title='Test Bug', got '%s'", voItem.Title)
	}
	if voItem.Status != "active" {
		t.Errorf("expected VO Status='active', got '%s'", voItem.Status)
	}

	// Verify type assertion works
	var _ vo.BugVO = voItem
}

// TestFilterBugChain_Severity（v1.5.0 MCP 扩面配套）：severity 在模型里是
// interface{}（JSON 数字/字符串双形态），过滤须归一后比较；0=不过滤。
func TestFilterBugChain_Severity(t *testing.T) {
	bugs := []zentao.Bug{
		{ID: 1, Severity: float64(2), AssignedTo: zentao.UserRef{Account: "zhangyi01"}},
		{ID: 2, Severity: "2", AssignedTo: zentao.UserRef{Account: "zhangyi01"}},
		{ID: 3, Severity: float64(3), AssignedTo: zentao.UserRef{Account: "zhangyi01"}},
		{ID: 4, Severity: nil, AssignedTo: zentao.UserRef{Account: "zhangyi01"}},
	}
	query := &dto.BugQueryDTO{Severity: 2}
	got := filterBugChain(bugs, query).Result()
	if len(got) != 2 || got[0].ID != 1 || got[1].ID != 2 {
		t.Fatalf("severity=2 应命中数字与字符串双形态两条，实得 %d 条", len(got))
	}

	query = &dto.BugQueryDTO{}
	if got = filterBugChain(bugs, query).Result(); len(got) != 4 {
		t.Fatalf("severity=0 应不过滤，实得 %d 条", len(got))
	}
}

// TestBugService_LiteSteps（v1.5.0）：Lite 列表剥 HTML 标签+折叠空白，超 300 字符
// 截断带指引标记——MCP 列表面不再背完整富文本（实弹 689KB 截断根因）。
func TestBugService_LiteSteps(t *testing.T) {
	long := "<p>" + string([]rune("重现步骤一二三四五六七八九十")) + "<br/></p><pre>" +
		string([]rune("长堆栈信息")) + " a=b</pre>"
	for len([]rune(long)) < 500 {
		long += "补"
	}
	bug := zentao.Bug{ID: 1, Title: "t", Steps: long}
	service := &BugService{client: nil}

	full := service.convertToVO([]zentao.Bug{bug})
	if full[0].Steps != long {
		t.Fatalf("非 Lite 形态 Steps 应原样保留")
	}

	list := service.convertToVO([]zentao.Bug{bug})
	applyLiteSteps(list)
	lite := list[0].Steps
	if strings.Contains(lite, "<p>") || strings.Contains(lite, "<br/>") {
		t.Fatalf("Lite Steps 应剥 HTML 标签: %q", lite[:80])
	}
	if len([]rune(lite)) > 340 {
		t.Fatalf("Lite Steps 应截断至 ~300 字: %d", len([]rune(lite)))
	}
	if !strings.Contains(lite, "get_bug") {
		t.Fatalf("Lite 截断应带 get_bug 指引标记: %q", lite[len(lite)-40:])
	}

	short := service.convertToVO([]zentao.Bug{zentao.Bug{ID: 2, Steps: "<p>第一步 打开页面</p>"}})
	applyLiteSteps(short)
	if short[0].Steps != "第一步 打开页面" {
		t.Fatalf("短 steps 应剥标签保留全文: %q", short[0].Steps)
	}
}
