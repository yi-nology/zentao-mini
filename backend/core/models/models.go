package models

// Response 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Product 产品模型
type Product struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Desc   string `json:"desc"`
}

// Project 项目模型
type Project struct {
	ID       int         `json:"id"`
	Name     string      `json:"name"`
	Code     string      `json:"code"`
	Model    string      `json:"model"`
	Type     string      `json:"type"`
	Status   string      `json:"status"`
	PM       interface{} `json:"pm"`
	Begin    string      `json:"begin"`
	End      string      `json:"end"`
	Progress interface{} `json:"progress"`
}

// Bug Bug模型
type Bug struct {
	ID            int         `json:"id"`
	Project       int         `json:"project"`
	Product       int         `json:"product"`
	Title         string      `json:"title"`
	Keywords      string      `json:"keywords"`
	Severity      int         `json:"severity"`
	Pri           int         `json:"pri"`
	Type          string      `json:"type"`
	OS            string      `json:"os"`
	Browser       string      `json:"browser"`
	Hardware      string      `json:"hardware"`
	Steps         string      `json:"steps"`
	Status        string      `json:"status"`
	SubStatus     string      `json:"subStatus"`
	Color         string      `json:"color"`
	Confirmed     int         `json:"confirmed"`
	PlanTime      string      `json:"planTime"`
	OpenedBy      UserRef     `json:"openedBy"`
	OpenedDate    string      `json:"openedDate"`
	OpenedBuild   string      `json:"openedBuild"`
	AssignedTo    UserRef     `json:"assignedTo"`
	AssignedDate  string      `json:"assignedDate"`
	Deadline      interface{} `json:"deadline"`
	ResolvedBy    interface{} `json:"resolvedBy"`
	Resolution    string      `json:"resolution"`
	ResolvedBuild string      `json:"resolvedBuild"`
	ResolvedDate  interface{} `json:"resolvedDate"`
	ClosedBy      interface{} `json:"closedBy"`
	ClosedDate    interface{} `json:"closedDate"`
	StatusName    string      `json:"statusName"`
	LifeCycle     string      `json:"lifeCycle"`
}

// Story 需求模型
type Story struct {
	ID           int         `json:"id"`
	Product      int         `json:"product"`
	Module       int         `json:"module"`
	Plan         string      `json:"plan"`
	Source       string      `json:"source"`
	Title        string      `json:"title"`
	Spec         string      `json:"spec"`
	Verify       string      `json:"verify"`
	Type         string      `json:"type"`
	Status       string      `json:"status"`
	Stage        string      `json:"stage"`
	Pri          int         `json:"pri"`
	Estimate     float64     `json:"estimate"`
	Version      int         `json:"version"`
	OpenedBy     interface{} `json:"openedBy"`
	OpenedDate   string      `json:"openedDate"`
	AssignedTo   interface{} `json:"assignedTo"`
	AssignedDate string      `json:"assignedDate"`
	ClosedBy     interface{} `json:"closedBy"`
	ClosedDate   string      `json:"closedDate"`
	ClosedReason string      `json:"closedReason"`
}

// Task 任务模型
type Task struct {
	ID           int         `json:"id"`
	Project      int         `json:"project"`
	Execution    int         `json:"execution"`
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Pri          int         `json:"pri"`
	Status       string      `json:"status"`
	AssignedTo   UserRef     `json:"assignedTo"`
	EstStarted   string      `json:"estStarted"`
	Deadline     string      `json:"deadline"`
	Estimate     float64     `json:"estimate"`
	Consumed     float64     `json:"consumed"`
	Left         float64     `json:"left"`
	Desc         string      `json:"desc"`
	OpenedBy     interface{} `json:"openedBy"`
	OpenedDate   string      `json:"openedDate"`
	FinishedBy   interface{} `json:"finishedBy"`
	FinishedDate interface{} `json:"finishedDate"`
	ClosedBy     interface{} `json:"closedBy"`
	ClosedDate   interface{} `json:"closedDate"`
	StatusName   string      `json:"statusName"`
}

// UserRef 用户引用模型
type UserRef struct {
	ID       int    `json:"id"`
	Account  string `json:"account"`
	Avatar   string `json:"avatar"`
	Realname string `json:"realname"`
}

// BugSearchParams Bug搜索参数
type BugSearchParams struct {
	ProductID  int
	Status     string
	AssignedTo string
	Keyword    string
	Severity   int
	Pri        int
	Limit      int
	Page       int
}

// Execution 执行/迭代模型
type Execution struct {
	ID      int    `json:"id"`
	Project int    `json:"project"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	Status  string `json:"status"`
	Type    string `json:"type"`
	Begin   string `json:"begin"`
	End     string `json:"end"`
}

// User 用户模型
type User struct {
	ID       int    `json:"id"`
	Account  string `json:"account"`
	Realname string `json:"realname"`
}

// PaginatedResult 分页结果
type PaginatedResult struct {
	List  interface{} `json:"list"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}
