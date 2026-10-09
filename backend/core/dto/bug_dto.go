package dto

type BugQueryDTO struct {
	ProductID    int    `form:"productId" json:"productId"`
	ProjectID    int    `form:"projectId" json:"projectId"`
	Status       string `form:"status" json:"status"`
	AssignedTo   string `form:"assignedTo" json:"assignedTo"`
	Version      string `form:"version" json:"version"`
	Type         string `form:"type" json:"type"`
	StartDate    string `form:"startDate" json:"startDate"`
	EndDate      string `form:"endDate" json:"endDate"`
	SpecificDate string `form:"specificDate" json:"specificDate"`
	// ResolvedStartDate/ResolvedEndDate 解决日期窗（扁鹊批次二百三十四挂账清偿：
	// 「今天解决了多少 bug」要按 resolvedDate 而非 openedDate 过滤——StartDate/EndDate
	// 只作用于 OpenedDate，本对独立作用于 ResolvedDate；任一非空即过滤，空=不约束）。
	ResolvedStartDate string `form:"resolvedStartDate" json:"resolvedStartDate"`
	ResolvedEndDate   string `form:"resolvedEndDate" json:"resolvedEndDate"`
	// ClosedStartDate/ClosedEndDate 关闭日期窗（v1.6.1，215 实弹 sess-1009-zsyb6euc
	// 专家拍板点：736 条已关闭 Bug 无 closedDate 过滤参数，今日关闭数只能全量翻页）。
	ClosedStartDate string `form:"closedStartDate" json:"closedStartDate"`
	ClosedEndDate   string `form:"closedEndDate" json:"closedEndDate"`
	Severity        int    `form:"severity" json:"severity"` // 严重度 1-5（0=不过滤；禅道口径 1致命~5建议）
	Page            int    `form:"page" json:"page"`
	PageSize        int    `form:"pageSize" json:"pageSize"`
	Lite            bool   `form:"lite" json:"lite"` // 轻量形态：列表 Steps 剥标签截断（MCP 面缺省开；Web API 缺省关零漂移）
}

func (dto *BugQueryDTO) Validate() error {
	if dto.Page <= 0 {
		dto.Page = 1
	}
	if dto.PageSize <= 0 {
		dto.PageSize = 20
	}
	if dto.PageSize > MaxPageSize {
		dto.PageSize = MaxPageSize
	}
	return nil
}
