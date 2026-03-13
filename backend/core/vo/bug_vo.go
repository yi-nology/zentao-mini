package vo

// BugVO Bug响应模型
// 用于API响应的数据结构
type BugVO struct {
	ID            int         `json:"id"`            // Bug ID
	Project       int         `json:"project"`       // 项目ID
	Product       int         `json:"product"`       // 产品ID
	Title         string      `json:"title"`         // 标题
	Keywords      string      `json:"keywords"`      // 关键词
	Severity      int         `json:"severity"`      // 严重程度
	Pri           int         `json:"pri"`           // 优先级
	Type          string      `json:"type"`          // 类型
	OS            string      `json:"os"`            // 操作系统
	Browser       string      `json:"browser"`       // 浏览器
	Hardware      string      `json:"hardware"`      // 硬件
	Steps         string      `json:"steps"`         // 重现步骤
	Status        string      `json:"status"`        // 状态
	SubStatus     string      `json:"subStatus"`     // 子状态
	Color         string      `json:"color"`         // 颜色
	Confirmed     int         `json:"confirmed"`     // 是否确认
	PlanTime      string      `json:"planTime"`      // 计划时间
	OpenedBy      UserRefVO   `json:"openedBy"`      // 创建人
	OpenedDate    string      `json:"openedDate"`    // 创建日期
	OpenedBuild   string      `json:"openedBuild"`   // 创建版本
	AssignedTo    UserRefVO   `json:"assignedTo"`    // 指派给
	AssignedDate  string      `json:"assignedDate"`  // 指派日期
	Deadline      interface{} `json:"deadline"`      // 截止日期
	ResolvedBy    interface{} `json:"resolvedBy"`    // 解决人
	Resolution    string      `json:"resolution"`    // 解决方案
	ResolvedBuild string      `json:"resolvedBuild"` // 解决版本
	ResolvedDate  interface{} `json:"resolvedDate"`  // 解决日期
	ClosedBy      interface{} `json:"closedBy"`      // 关闭人
	ClosedDate    interface{} `json:"closedDate"`    // 关闭日期
	StatusName    string      `json:"statusName"`    // 状态名称
	LifeCycle     string      `json:"lifeCycle"`     // 生命周期
}

// UserRefVO 用户引用响应模型
type UserRefVO struct {
	ID       int    `json:"id"`       // 用户ID
	Account  string `json:"account"`  // 账号
	Avatar   string `json:"avatar"`   // 头像
	Realname string `json:"realname"` // 真实姓名
}
