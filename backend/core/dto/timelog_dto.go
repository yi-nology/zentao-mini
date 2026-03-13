package dto

// TimelogQueryDTO 工时统计查询请求参数
// 使用驼峰命名风格（camelCase）
type TimelogQueryDTO struct {
	ProductID   string `form:"productId" json:"productId"`     // 产品ID，必需
	ProjectID   string `form:"projectId" json:"projectId"`     // 项目ID
	ExecutionID string `form:"executionId" json:"executionId"` // 执行ID
	AssignedTo  string `form:"assignedTo" json:"assignedTo"`   // 指派人
	DateFrom    string `form:"dateFrom" json:"dateFrom"`       // 开始日期(YYYY-MM-DD)，必需
	DateTo      string `form:"dateTo" json:"dateTo"`           // 结束日期(YYYY-MM-DD)，必需
}

// Validate 验证参数
func (dto *TimelogQueryDTO) Validate() error {
	// 这里可以添加更详细的验证逻辑
	return nil
}

// IsEmpty 检查是否为空
func (dto *TimelogQueryDTO) IsEmpty() bool {
	return dto.ProductID == ""
}
