package zentao

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yi-nology/common/biz/zentao"
)

// Client 封装禅道 SDK 客户端，支持 Token 缓存
type Client struct {
	sdkClient      *zentao.Client
	account        string
	password       string
	server         string
	token          string
	tokenExpiry    time.Time
	usersCache     []zentao.User
	usersExpiry    time.Time
	productsCache  []zentao.Product
	productsExpiry time.Time
	mu             sync.RWMutex
}

// NewClient 创建新的禅道客户端
func NewClient(server, account, password string) *Client {
	// 确保server URL格式正确
	if !strings.HasPrefix(server, "http://") && !strings.HasPrefix(server, "https://") {
		server = "http://" + server
	}
	if strings.HasSuffix(server, "/") {
		server = strings.TrimSuffix(server, "/")
	}

	sdkClient := zentao.NewClient(server)
	// 设置更长的超时时间，避免复杂请求超时
	sdkClient.SetTimeout(120 * time.Second) // 增加到120秒，确保复杂查询不会超时
	client := &Client{
		sdkClient: sdkClient,
		account:   account,
		password:  password,
		server:    server,
	}
	// 启动Token自动刷新机制
	go client.startTokenRefreshTask()
	return client
}

// startTokenRefreshTask 启动Token自动刷新任务
func (c *Client) startTokenRefreshTask() {
	ticker := time.NewTicker(2 * time.Hour) // 每12小时检查一次
	defer ticker.Stop()

	for {
		<-ticker.C
		// 检查Token是否即将过期（剩余时间小于2小时）
		c.mu.RLock()
		remaining := c.tokenExpiry.Sub(time.Now())
		c.mu.RUnlock()

		if remaining < 2*time.Hour {
			// 刷新Token
			if _, err := c.RefreshToken(); err != nil {
				// 刷新失败，下次再试
				continue
			}
		}
	}
}

// IsTokenExpired 检查 Token 是否已过期
func (c *Client) IsTokenExpired() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token == "" || time.Now().After(c.tokenExpiry)
}

// getToken 获取有效的 Token（带缓存）
func (c *Client) getToken() (string, error) {
	c.mu.RLock()
	if c.token != "" && time.Now().Before(c.tokenExpiry) {
		token := c.token
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// 双重检查
	if c.token != "" && time.Now().Before(c.tokenExpiry) {
		return c.token, nil
	}

	// 尝试获取Token，最多重试3次
	var token string
	var err error
	for i := 0; i < 3; i++ {
		token, err = c.sdkClient.GetToken(c.account, c.password)
		if err == nil {
			break
		}
		// 重试前等待一段时间
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if err != nil {
		return "", err
	}

	c.token = token
	// Token 有效期设置为 23 小时（保险起见）
	c.tokenExpiry = time.Now().Add(23 * time.Hour)
	c.sdkClient.SetToken(token)

	return token, nil
}

// RefreshToken 强制刷新 Token
func (c *Client) RefreshToken() (string, error) {

	// 尝试获取Token，最多重试3次
	var token string
	var err error
	for i := 0; i < 3; i++ {
		token, err = c.sdkClient.GetToken(c.account, c.password)
		if err == nil {
			break
		}
		// 重试前等待一段时间
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if err != nil {
		return "", err
	}

	c.token = token
	// Token 有效期设置为 23 小时（保险起见）
	c.tokenExpiry = time.Now().Add(23 * time.Hour)
	c.sdkClient.SetToken(token)

	return token, nil
}

// UpdateConfig 更新客户端配置并刷新Token
func (c *Client) UpdateConfig(server, account, password string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 确保server URL格式正确
	if !strings.HasPrefix(server, "http://") && !strings.HasPrefix(server, "https://") {
		server = "http://" + server
	}
	if strings.HasSuffix(server, "/") {
		server = strings.TrimSuffix(server, "/")
	}

	// 更新配置
	c.server = server
	c.account = account
	c.password = password

	// 重新创建SDK客户端
	c.sdkClient = zentao.NewClient(server)
	c.sdkClient.SetTimeout(10 * time.Second)

	// 清除缓存
	c.token = ""
	c.tokenExpiry = time.Time{}
	c.usersCache = nil
	c.usersExpiry = time.Time{}
	c.productsCache = nil
	c.productsExpiry = time.Time{}

	// 立即刷新Token
	_, err := c.RefreshToken()
	return err
}

// GetAccount 获取当前登录用户的账号
func (c *Client) GetAccount() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.account
}

// GetProducts 获取产品列表
func (c *Client) GetProducts() ([]zentao.Product, error) {
	// 检查缓存
	c.mu.RLock()
	if len(c.productsCache) > 0 && time.Now().Before(c.productsExpiry) {
		// 缓存有效，直接返回
		products := c.productsCache
		c.mu.RUnlock()
		return products, nil
	}
	c.mu.RUnlock()

	// 缓存无效，重新获取所有产品
	if _, err := c.getToken(); err != nil {
		return nil, err
	}

	// 先获取第一页，获取总数
	firstPageResponse, err := c.sdkClient.GetProducts(1, 100)
	if err != nil {
		return nil, err
	}

	var allProducts []zentao.Product
	allProducts = append(allProducts, firstPageResponse.Products...)

	// 计算总页数
	total := firstPageResponse.Total
	pageSize := 100
	totalPages := (total + pageSize - 1) / pageSize

	// 如果有更多页，继续获取
	if totalPages > 1 {
		for page := 2; page <= totalPages; page++ {
			response, err := c.sdkClient.GetProducts(page, pageSize)
			if err != nil {
				return nil, err
			}
			allProducts = append(allProducts, response.Products...)
		}
	}

	// 更新缓存
	c.mu.Lock()
	c.productsCache = allProducts
	c.productsExpiry = time.Now().Add(24 * time.Hour) // 24小时过期
	c.mu.Unlock()

	return allProducts, nil
}

// GetProduct 获取产品详情
func (c *Client) GetProduct(productID int) (*zentao.Product, error) {
	if _, err := c.getToken(); err != nil {
		return nil, err
	}
	return c.sdkClient.GetProduct(productID)
}

// GetAllProjects 获取所有项目列表
func (c *Client) GetAllProjects(limit int) ([]zentao.Project, error) {
	if _, err := c.getToken(); err != nil {
		return nil, err
	}
	response, err := c.sdkClient.GetAllProjects(1, limit)
	if err != nil {
		return nil, err
	}
	return response.Projects, nil
}

// GetProjectsByProduct 获取产品关联的项目列表
func (c *Client) GetProjectsByProduct(productID int) ([]zentao.Project, error) {
	if _, err := c.getToken(); err != nil {
		return nil, err
	}
	response, err := c.sdkClient.GetProjectsByProduct(productID, 1, 100)
	if err != nil {
		return nil, err
	}
	return response.Projects, nil
}

// GetProject 获取项目详情
func (c *Client) GetProject(projectID int) (*zentao.Project, error) {
	if _, err := c.getToken(); err != nil {
		return nil, err
	}
	return c.sdkClient.GetProject(projectID)
}

// GetBugs 获取产品的 Bug 列表
func (c *Client) GetBugs(productID int, limit int) ([]zentao.Bug, error) {
	if _, err := c.getToken(); err != nil {
		return nil, err
	}
	response, err := c.sdkClient.GetBugs(productID, 1, limit)
	if err != nil {
		return nil, err
	}
	return response.Bugs, nil
}

// GetBugsByProject 根据项目 ID 过滤 Bug 列表
func (c *Client) GetBugsByProject(productID, projectID int, limit int) ([]zentao.Bug, error) {
	if _, err := c.getToken(); err != nil {
		return nil, err
	}
	response, err := c.sdkClient.GetBugsByProject(productID, projectID, 1, limit)
	if err != nil {
		return nil, err
	}
	return response.Bugs, nil
}

// GetBugsByStatus 根据状态过滤 Bug 列表
func (c *Client) GetBugsByStatus(productID int, status string, limit int) ([]zentao.Bug, error) {
	if _, err := c.getToken(); err != nil {
		return nil, err
	}
	response, err := c.sdkClient.GetBugsByStatus(productID, status, 1, limit)
	if err != nil {
		return nil, err
	}
	return response.Bugs, nil
}

// SearchBugs 搜索 Bug（支持多条件过滤）
func (c *Client) SearchBugs(params zentao.BugSearchParams) ([]zentao.Bug, error) {
	if _, err := c.getToken(); err != nil {
		return nil, err
	}
	response, err := c.sdkClient.SearchBugs(params)
	if err != nil {
		return nil, err
	}
	return response.Bugs, nil
}

// GetBug 获取 Bug 详情
func (c *Client) GetBug(bugID int) (*zentao.Bug, error) {
	if _, err := c.getToken(); err != nil {
		return nil, err
	}
	return c.sdkClient.GetBug(bugID)
}

// GetStoriesByProduct 获取产品的需求列表
func (c *Client) GetStoriesByProduct(productID int, limit int) ([]zentao.Story, error) {
	if _, err := c.getToken(); err != nil {
		return nil, err
	}
	response, err := c.sdkClient.GetStoriesByProduct(productID, 1, limit)
	if err != nil {
		return nil, err
	}
	return response.Stories, nil
}

// GetStoriesByProject 获取项目的需求列表
func (c *Client) GetStoriesByProject(projectID int, limit int) ([]zentao.Story, error) {
	if _, err := c.getToken(); err != nil {
		return nil, err
	}
	response, err := c.sdkClient.GetStoriesByProject(projectID, 1, limit)
	if err != nil {
		return nil, err
	}
	return response.Stories, nil
}

// GetStoriesByExecution 获取执行的需求列表
func (c *Client) GetStoriesByExecution(executionID int, limit int) ([]zentao.Story, error) {
	if _, err := c.getToken(); err != nil {
		return nil, err
	}
	response, err := c.sdkClient.GetStoriesByExecution(executionID, 1, limit)
	if err != nil {
		return nil, err
	}
	return response.Stories, nil
}

// GetStory 获取需求详情
func (c *Client) GetStory(storyID int) (*zentao.Story, error) {
	if _, err := c.getToken(); err != nil {
		return nil, err
	}
	return c.sdkClient.GetStory(storyID)
}

// GetTasks 获取执行的任务列表
func (c *Client) GetTasks(executionID int, limit int) ([]zentao.Task, error) {
	if _, err := c.getToken(); err != nil {
		return nil, err
	}
	response, err := c.sdkClient.GetTasks(executionID, 1, limit)
	if err != nil {
		return nil, err
	}
	return response.Tasks, nil
}

// GetTask 获取任务详情
func (c *Client) GetTask(taskID int) (*zentao.Task, error) {
	if _, err := c.getToken(); err != nil {
		return nil, err
	}
	return c.sdkClient.GetTask(taskID)
}

// GetExecutions 获取执行列表
func (c *Client) GetExecutions(projectID int) ([]zentao.Execution, error) {
	if _, err := c.getToken(); err != nil {
		return nil, err
	}
	response, err := c.sdkClient.GetExecutions(projectID, 1, 100)
	if err != nil {
		return nil, err
	}
	return response.Executions, nil
}

// GetUsers 获取用户列表（支持分页）
func (c *Client) GetUsers(page, limit int) (*zentao.UserListResponse, error) {
	// 检查缓存
	c.mu.RLock()
	if len(c.usersCache) > 0 && time.Now().Before(c.usersExpiry) {
		// 缓存有效，直接返回
		users := c.usersCache
		c.mu.RUnlock()

		// 计算分页
		total := len(users)
		start := (page - 1) * limit
		end := start + limit
		if start >= total {
			return &zentao.UserListResponse{
				Users: []zentao.User{},
				Page:  page,
				Total: total,
				Limit: limit,
			}, nil
		}
		if end > total {
			end = total
		}

		return &zentao.UserListResponse{
			Users: users[start:end],
			Page:  page,
			Total: total,
			Limit: limit,
		}, nil
	}
	c.mu.RUnlock()

	// 缓存无效，重新获取所有用户
	if _, err := c.getToken(); err != nil {
		return nil, err
	}

	// 分页获取所有用户，每次100人
	var allUsers []zentao.User
	currentPage := 1
	pageSize := 100

	for {
		response, err := c.sdkClient.GetUsers(currentPage, pageSize)
		if err != nil {
			return nil, err
		}

		allUsers = append(allUsers, response.Users...)

		// 检查是否还有更多数据
		if len(response.Users) < pageSize {
			break
		}

		currentPage++
	}

	// 更新缓存
	c.mu.Lock()
	c.usersCache = allUsers
	c.usersExpiry = time.Now().Add(24 * time.Hour) // 24小时过期
	c.mu.Unlock()

	// 计算分页
	total := len(allUsers)
	start := (page - 1) * limit
	end := start + limit
	if start >= total {
		return &zentao.UserListResponse{
			Users: []zentao.User{},
			Page:  page,
			Total: total,
			Limit: limit,
		}, nil
	}
	if end > total {
		end = total
	}

	return &zentao.UserListResponse{
		Users: allUsers[start:end],
		Page:  page,
		Total: total,
		Limit: limit,
	}, nil
}

// GetUsersAll 获取所有用户列表
func (c *Client) GetUsersAll() ([]zentao.User, error) {
	// 检查缓存
	c.mu.RLock()
	if len(c.usersCache) > 0 && time.Now().Before(c.usersExpiry) {
		// 缓存有效，直接返回
		users := c.usersCache
		c.mu.RUnlock()
		return users, nil
	}
	c.mu.RUnlock()

	// 缓存无效，重新获取所有用户
	if _, err := c.getToken(); err != nil {
		return nil, err
	}

	// 分页获取所有用户，每次100人
	var allUsers []zentao.User
	currentPage := 1
	pageSize := 100

	for {
		response, err := c.sdkClient.GetUsers(currentPage, pageSize)
		if err != nil {
			return nil, err
		}

		allUsers = append(allUsers, response.Users...)

		// 检查是否还有更多数据
		if len(response.Users) < pageSize {
			break
		}

		currentPage++
	}

	// 更新缓存
	c.mu.Lock()
	c.usersCache = allUsers
	c.usersExpiry = time.Now().Add(24 * time.Hour) // 24小时过期
	c.mu.Unlock()

	return allUsers, nil
}

// effortItem 工时记录项
type effortItem struct {
	ID        int     `json:"id"`
	TaskID    int     `json:"taskId"`
	TaskName  string  `json:"taskName"`
	TaskType  string  `json:"taskType"`
	Product   string  `json:"product"`
	Project   string  `json:"project"`
	Execution string  `json:"execution"`
	Account   string  `json:"account"`
	Date      string  `json:"date"`
	Consumed  float64 `json:"consumed"`
	Left      float64 `json:"left"`
	Work      string  `json:"work"`
}

// statItem 统计项
type statItem struct {
	Name  string  `json:"name"`
	Hours float64 `json:"hours"`
	Count int     `json:"count"`
}

// dailyStat 每日统计
type dailyStat struct {
	Date  string  `json:"date"`
	Hours float64 `json:"hours"`
	Count int     `json:"count"`
}

// mapToSlice 将统计映射转换为切片
func mapToSlice(m map[string]*statItem) []statItem {
	result := make([]statItem, 0, len(m))
	for _, v := range m {
		result = append(result, *v)
	}
	return result
}

// GetTimelogAnalysis 获取工时统计分析
func (c *Client) GetTimelogAnalysis(productID, projectID, executionID, assignedTo, dateFrom, dateTo string) (map[string]interface{}, error) {
	if _, err := c.getToken(); err != nil {
		return nil, err
	}

	// 转换产品ID为整数
	prodID, err := strconv.Atoi(productID)
	if err != nil {
		return nil, fmt.Errorf("无效的productId")
	}

	// 转换项目ID和执行ID为整数
	var filterProjectID, filterExecutionID int
	if projectID != "" {
		filterProjectID, _ = strconv.Atoi(projectID)
	}
	if executionID != "" {
		filterExecutionID, _ = strconv.Atoi(executionID)
	}

	// 获取产品名称
	productName := fmt.Sprintf("产品%d", prodID)
	// 只获取需要的产品信息，避免获取所有产品
	product, err := c.sdkClient.GetProduct(prodID)
	if err == nil && product != nil {
		productName = product.Name
	}

	// 获取产品下的项目列表
	projectsResponse, err := c.sdkClient.GetProjectsByProduct(prodID, 1, 100)
	if err != nil {
		return nil, fmt.Errorf("获取项目列表失败: %w", err)
	}
	projects := projectsResponse.Projects

	// 如果指定了项目，只取该项目
	if filterProjectID > 0 {
		filtered := make([]zentao.Project, 0)
		for _, p := range projects {
			if p.ID == filterProjectID {
				filtered = append(filtered, p)
				break
			}
		}
		projects = filtered
	}

	// 收集执行列表
	type execWithContext struct {
		Exec     zentao.Execution
		ProjName string
	}
	var allExecs []execWithContext
	var mu sync.Mutex
	var wg sync.WaitGroup
	// 限制并发数，避免过多请求导致超时
	execSem := make(chan struct{}, 3) // 最多3个并发请求，减少服务器压力

	if filterExecutionID > 0 {
		// 直接使用指定的执行
		projName := ""
		for _, p := range projects {
			projName = p.Name
			break
		}
		allExecs = append(allExecs, execWithContext{
			Exec:     zentao.Execution{ID: filterExecutionID},
			ProjName: projName,
		})
	} else {
		for _, proj := range projects {
			wg.Add(1)
			go func(p zentao.Project) {
				defer wg.Done()
				execSem <- struct{}{}
				defer func() { <-execSem }()
				execsResponse, err := c.sdkClient.GetExecutions(p.ID, 1, 100)
				if err != nil {
					// 记录错误但不中断整体流程
					return
				}
				execs := execsResponse.Executions
				mu.Lock()
				for _, e := range execs {
					allExecs = append(allExecs, execWithContext{Exec: e, ProjName: p.Name})
				}
				mu.Unlock()
			}(proj)
		}
		wg.Wait()
	}

	// 并发收集所有任务（含基本信息用于关联）
	type taskContext struct {
		Task     zentao.Task
		ProjName string
		ExecName string
	}
	var allTaskCtx []taskContext
	taskSem := make(chan struct{}, 3) // 最多3个并发请求

	for _, ec := range allExecs {
		wg.Add(1)
		go func(e execWithContext) {
			defer wg.Done()
			taskSem <- struct{}{}
			defer func() { <-taskSem }()
			tasksResponse, err := c.sdkClient.GetTasks(e.Exec.ID, 1, 500)
			if err != nil {
				// 记录错误但不中断整体流程
				return
			}
			tasks := tasksResponse.Tasks
			mu.Lock()
			for _, t := range tasks {
				if t.Consumed <= 0 {
					continue
				}
				allTaskCtx = append(allTaskCtx, taskContext{Task: t, ProjName: e.ProjName, ExecName: e.Exec.Name})
			}
			mu.Unlock()
		}(ec)
	}
	wg.Wait()

	// 并发获取每个任务的 effort 日志
	var allEfforts []effortItem
	sem := make(chan struct{}, 5) // 限制并发数为5，避免过多请求

	for _, tc := range allTaskCtx {
		wg.Add(1)
		go func(tc taskContext) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			efforts, err := c.sdkClient.GetTaskEfforts(tc.Task.ID)
			if err != nil {
				// 记录错误但不中断整体流程
				return
			}
			mu.Lock()
			for _, e := range efforts {
				// 用户过滤（effort 记录人）
				if assignedTo != "" && e.Account != assignedTo {
					continue
				}
				// 日期范围过滤
				if dateFrom != "" && e.Date < dateFrom {
					continue
				}
				if dateTo != "" && e.Date > dateTo {
					continue
				}

				allEfforts = append(allEfforts, effortItem{
					ID:        e.ID,
					TaskID:    tc.Task.ID,
					TaskName:  tc.Task.Name,
					TaskType:  tc.Task.Type,
					Product:   productName,
					Project:   tc.ProjName,
					Execution: tc.ExecName,
					Account:   e.Account,
					Date:      e.Date,
					Consumed:  e.Consumed,
					Left:      e.Left,
					Work:      e.Work,
				})
			}
			mu.Unlock()
		}(tc)
	}
	wg.Wait()

	// 聚合统计
	var totalHours float64
	byProjectMap := make(map[string]*statItem)
	byTypeMap := make(map[string]*statItem)
	byDateMap := make(map[string]*dailyStat)

	typeNames := map[string]string{
		"devel": "开发", "design": "设计", "test": "测试",
		"study": "研究", "discuss": "讨论", "ui": "界面",
		"affair": "事务", "misc": "其他",
	}

	for _, e := range allEfforts {
		totalHours += e.Consumed

		// 按项目
		if _, ok := byProjectMap[e.Project]; !ok {
			byProjectMap[e.Project] = &statItem{Name: e.Project}
		}
		byProjectMap[e.Project].Hours += e.Consumed
		byProjectMap[e.Project].Count++

		// 按类型
		typeName := e.TaskType
		if tn, ok := typeNames[e.TaskType]; ok {
			typeName = tn
		}
		if _, ok := byTypeMap[typeName]; !ok {
			byTypeMap[typeName] = &statItem{Name: typeName}
		}
		byTypeMap[typeName].Hours += e.Consumed
		byTypeMap[typeName].Count++

		// 按日期
		if _, ok := byDateMap[e.Date]; !ok {
			byDateMap[e.Date] = &dailyStat{Date: e.Date}
		}
		byDateMap[e.Date].Hours += e.Consumed
		byDateMap[e.Date].Count++
	}

	byProject := mapToSlice(byProjectMap)
	byType := mapToSlice(byTypeMap)

	// 日期统计转slice并排序
	byDate := make([]dailyStat, 0, len(byDateMap))
	for _, v := range byDateMap {
		byDate = append(byDate, *v)
	}
	// 按日期排序
	for i := 0; i < len(byDate); i++ {
		for j := i + 1; j < len(byDate); j++ {
			if byDate[i].Date > byDate[j].Date {
				byDate[i], byDate[j] = byDate[j], byDate[i]
			}
		}
	}

	// 准备返回数据
	effortsData := make([]map[string]interface{}, len(allEfforts))
	for i, e := range allEfforts {
		effortsData[i] = map[string]interface{}{
			"id":        e.ID,
			"taskId":    e.TaskID,
			"taskName":  e.TaskName,
			"taskType":  e.TaskType,
			"product":   e.Product,
			"project":   e.Project,
			"execution": e.Execution,
			"account":   e.Account,
			"date":      e.Date,
			"consumed":  e.Consumed,
			"left":      e.Left,
			"work":      e.Work,
		}
	}

	return map[string]interface{}{
		"totalHours":  totalHours,
		"effortCount": len(allEfforts),
		"taskCount":   len(allTaskCtx),
		"byProject":   byProject,
		"byType":      byType,
		"byDate":      byDate,
		"efforts":     effortsData,
	}, nil
}
