package dto

// BugQueryDTO Bug查询请求参数
// 使用驼峰命名风格（camelCase）
type BugQueryDTO struct {
	ProductID    int    `form:"productId" json:"productId"`       // 产品ID
	ProjectID    int    `form:"projectId" json:"projectId"`       // 项目ID
	Status       string `form:"status" json:"status"`             // 状态(active, resolved, closed等)
	AssignedTo   string `form:"assignedTo" json:"assignedTo"`     // 指派人账号
	StartDate    string `form:"startDate" json:"startDate"`       // 开始日期(YYYY-MM-DD)
	EndDate      string `form:"endDate" json:"endDate"`           // 结束日期(YYYY-MM-DD)
	SpecificDate string `form:"specificDate" json:"specificDate"` // 具体日期(YYYY-MM-DD)
	Page         int    `form:"page" json:"page"`                 // 页码，默认1
	Limit        int    `form:"limit" json:"limit"`               // 每页数量，默认20
}

// Validate 验证参数
func (dto *BugQueryDTO) Validate() error {
	// 设置默认值
	if dto.Page <= 0 {
		dto.Page = 1
	}
	if dto.Limit <= 0 {
		dto.Limit = 20
	}
	return nil
}
