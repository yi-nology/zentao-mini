package dto

// ProjectQueryDTO 项目查询请求参数
type ProjectQueryDTO struct {
	ProductID int `form:"productId" json:"productId"` // 产品ID（可选）
}

// ExecutionQueryDTO 执行/迭代查询请求参数
type ExecutionQueryDTO struct {
	ProjectID int `form:"projectId" json:"projectId"` // 项目ID（可选）
	ProductID int `form:"productId" json:"productId"` // 产品ID（可选）
}
