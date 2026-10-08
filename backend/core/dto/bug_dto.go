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
	Severity     int    `form:"severity" json:"severity"` // 严重度 1-5（0=不过滤；禅道口径 1致命~5建议）
	Page         int    `form:"page" json:"page"`
	PageSize     int    `form:"pageSize" json:"pageSize"`
	Lite         bool   `form:"lite" json:"lite"` // 轻量形态：列表 Steps 剥标签截断（MCP 面缺省开；Web API 缺省关零漂移）
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
