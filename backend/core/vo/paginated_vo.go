package vo

// PaginatedVO 分页响应模型
// 用于API响应的分页数据结构
type PaginatedVO struct {
	List  interface{} `json:"list"`  // 数据列表
	Total int         `json:"total"` // 总数
	Page  int         `json:"page"`  // 当前页
	Limit int         `json:"limit"` // 每页数量
}
