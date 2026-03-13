package zentao

import (
	"chandao-mini/backend/core/logger"
	"chandao-mini/backend/core/metrics"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yi-nology/common/biz/zentao"
	"go.uber.org/zap"
)

// Client 封装禅道 SDK 客户端，支持 Token 缓存
type Client struct {
	sdkClient      *zentao.Client
	account        string
	password       *SecureString // 使用安全字符串存储密码
	server         string
	token          *SecureString // 使用安全字符串存储token
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
		password:  NewSecureString(password), // 使用安全字符串存储密码
		server:    server,
		token:     NewSecureString(""), // 初始化空的安全字符串
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
	return c.token.Get() == "" || time.Now().After(c.tokenExpiry)
}

// getToken 获取有效的 Token（带缓存）
func (c *Client) getToken() (string, error) {
	c.mu.RLock()
	tokenStr := c.token.Get()
	if tokenStr != "" && time.Now().Before(c.tokenExpiry) {
		c.mu.RUnlock()
		// 缓存命中
		metrics.RecordCacheHit("token")
		return tokenStr, nil
	}
	c.mu.RUnlock()

	// 缓存未命中
	metrics.RecordCacheMiss("token")
	start := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	// 双重检查
	tokenStr = c.token.Get()
	if tokenStr != "" && time.Now().Before(c.tokenExpiry) {
		return tokenStr, nil
	}

	// 尝试获取Token，最多重试3次
	var token string
	var err error
	passwordStr := c.password.Get() // 临时获取密码
	for i := 0; i < 3; i++ {
		token, err = c.sdkClient.GetToken(c.account, passwordStr)
		if err == nil {
			break
		}
		// 重试前等待一段时间
		time.Sleep(time.Duration(i+1) * time.Second)
		logger.Warn("Failed to get token, retrying",
			zap.Int("attempt", i+1),
			zap.Error(err),
		)
	}
	// 清除临时密码变量
	passwordStr = ""

	if err != nil {
		logger.Error("Failed to get token after retries", zap.Error(err))
		return "", err
	}

	c.token.Set(token) // 使用安全字符串存储token
	// Token 有效期设置为 23 小时（保险起见）
	c.tokenExpiry = time.Now().Add(23 * time.Hour)
	c.sdkClient.SetToken(token)

	duration := time.Since(start)
	metrics.RecordCacheOperation("token", "refresh", duration)
	metrics.RecordTokenRefresh()

	logger.Info("Token refreshed successfully",
		zap.Duration("duration", duration),
		zap.Time("expiry", c.tokenExpiry),
	)

	return token, nil
}

// RefreshToken 强制刷新 Token
func (c *Client) RefreshToken() (string, error) {

	// 尝试获取Token，最多重试3次
	var token string
	var err error
	passwordStr := c.password.Get() // 临时获取密码
	for i := 0; i < 3; i++ {
		token, err = c.sdkClient.GetToken(c.account, passwordStr)
		if err == nil {
			break
		}
		// 重试前等待一段时间
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	// 清除临时密码变量
	passwordStr = ""

	if err != nil {
		return "", err
	}

	c.token.Set(token) // 使用安全字符串存储token
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
	c.password.Set(password) // 使用安全字符串存储密码

	// 重新创建SDK客户端
	c.sdkClient = zentao.NewClient(server)
	c.sdkClient.SetTimeout(10 * time.Second)

	// 清除缓存
	c.token.Set("") // 清除token
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
		metrics.RecordCacheHit("products")
		return products, nil
	}
	c.mu.RUnlock()

	// 缓存无效，重新获取所有产品
	metrics.RecordCacheMiss("products")
	start := time.Now()

	if _, err := c.getToken(); err != nil {
		return nil, err
	}

	// 先获取第一页，获取总数
	firstPageResponse, err := c.sdkClient.GetProducts(1, 100)
	if err != nil {
		metrics.RecordZentaoAPIRequest("products", "GET", time.Since(start), err)
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
				metrics.RecordZentaoAPIRequest("products", "GET", time.Since(start), err)
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

	duration := time.Since(start)
	metrics.RecordCacheOperation("products", "fetch", duration)
	metrics.RecordZentaoAPIRequest("products", "GET", duration, nil)

	logger.Info("Products fetched",
		zap.Int("count", len(allProducts)),
		zap.Duration("duration", duration),
	)

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
		metrics.RecordCacheHit("users")

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
	metrics.RecordCacheMiss("users")
	startTime := time.Now()

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
			metrics.RecordZentaoAPIRequest("users", "GET", time.Since(startTime), err)
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

	duration := time.Since(startTime)
	metrics.RecordCacheOperation("users", "fetch", duration)
	metrics.RecordZentaoAPIRequest("users", "GET", duration, nil)

	logger.Info("Users fetched",
		zap.Int("count", len(allUsers)),
		zap.Duration("duration", duration),
	)

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
// 使用WorkerPool进行并发控制，提升性能和可维护性
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

	// 定义执行上下文结构
	type execWithContext struct {
		Exec     zentao.Execution
		ProjName string
	}

	// 第一步：收集执行列表（使用WorkerPool）
	var allExecs []execWithContext
	var mu sync.Mutex

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
		// 使用WorkerPool并发获取执行列表
		execPool := NewWorkerPool(3, len(projects)) // 3个并发worker
		defer execPool.Shutdown()

		execTasks := make([]Task, len(projects))
		for i, proj := range projects {
			proj := proj // 捕获变量
			execTasks[i] = func() (interface{}, error) {
				execsResponse, err := c.sdkClient.GetExecutions(proj.ID, 1, 100)
				if err != nil {
					return nil, err
				}
				return execsResponse.Executions, nil
			}
		}

		// 处理结果
		for i, result := range execPool.ProcessBatch(execTasks) {
			if result.Error == nil && result.Value != nil {
				execs := result.Value.([]zentao.Execution)
				mu.Lock()
				for _, e := range execs {
					allExecs = append(allExecs, execWithContext{Exec: e, ProjName: projects[i].Name})
				}
				mu.Unlock()
			}
		}
	}

	// 第二步：收集任务列表（使用WorkerPool）
	type taskContext struct {
		Task     zentao.Task
		ProjName string
		ExecName string
	}
	var allTaskCtx []taskContext

	taskPool := NewWorkerPool(3, len(allExecs)) // 3个并发worker
	defer taskPool.Shutdown()

	taskTasks := make([]Task, len(allExecs))
	for i, ec := range allExecs {
		ec := ec // 捕获变量
		taskTasks[i] = func() (interface{}, error) {
			tasksResponse, err := c.sdkClient.GetTasks(ec.Exec.ID, 1, 500)
			if err != nil {
				return nil, err
			}
			// 过滤掉没有消耗工时的任务
			var filteredTasks []zentao.Task
			for _, t := range tasksResponse.Tasks {
				if t.Consumed > 0 {
					filteredTasks = append(filteredTasks, t)
				}
			}
			return struct {
				Tasks    []zentao.Task
				ProjName string
				ExecName string
			}{Tasks: filteredTasks, ProjName: ec.ProjName, ExecName: ec.Exec.Name}, nil
		}
	}

	// 处理结果
	for _, result := range taskPool.ProcessBatch(taskTasks) {
		if result.Error == nil && result.Value != nil {
			data := result.Value.(struct {
				Tasks    []zentao.Task
				ProjName string
				ExecName string
			})
			mu.Lock()
			for _, t := range data.Tasks {
				allTaskCtx = append(allTaskCtx, taskContext{
					Task:     t,
					ProjName: data.ProjName,
					ExecName: data.ExecName,
				})
			}
			mu.Unlock()
		}
	}

	// 第三步：获取工时记录（使用WorkerPool）
	var allEfforts []effortItem

	effortPool := NewWorkerPool(5, len(allTaskCtx)) // 5个并发worker
	defer effortPool.Shutdown()

	effortTasks := make([]Task, len(allTaskCtx))
	for i, tc := range allTaskCtx {
		tc := tc // 捕获变量
		effortTasks[i] = func() (interface{}, error) {
			efforts, err := c.sdkClient.GetTaskEfforts(tc.Task.ID)
			if err != nil {
				return nil, err
			}

			// 过滤工时记录
			var filteredEfforts []effortItem
			for _, e := range efforts {
				// 用户过滤
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

				filteredEfforts = append(filteredEfforts, effortItem{
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
			return filteredEfforts, nil
		}
	}

	// 处理结果
	for _, result := range effortPool.ProcessBatch(effortTasks) {
		if result.Error == nil && result.Value != nil {
			efforts := result.Value.([]effortItem)
			mu.Lock()
			allEfforts = append(allEfforts, efforts...)
			mu.Unlock()
		}
	}

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
