package zentao

import (
	"github.com/yi-nology/zentao-mini/backend/core/logger"
	"github.com/yi-nology/zentao-mini/backend/core/metrics"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yi-nology/common/biz/zentao"
	"go.uber.org/zap"
)

// Client 封装禅道 SDK 客户端，支持 Token 缓存
type Client struct {
	sdkClient   *zentao.Client
	account     string
	password    *SecureString
	server      string
	token       *SecureString
	tokenExpiry atomic.Int64
	mu          sync.RWMutex
	connected   atomic.Bool
	refreshing  atomic.Bool
}

const tokenTTL = 23 * time.Hour

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
	ticker := time.NewTicker(2 * time.Hour)
	defer ticker.Stop()

	for {
		<-ticker.C
		if c.isTokenExpired() {
			if _, err := c.RefreshToken(); err != nil {
				continue
			}
		}
	}
}

func (c *Client) isTokenExpired() bool {
	expiry := time.Unix(c.tokenExpiry.Load(), 0)
	return time.Now().After(expiry)
}

// IsTokenExpired 检查 Token 是否已过期
func (c *Client) IsTokenExpired() bool {
	return c.isTokenExpired()
}

// getToken 获取有效的 Token（带缓存）
// 核心设计：网络 I/O 在锁外执行，绝不持锁做 HTTP 请求
func (c *Client) getToken() (string, error) {
	tokenStr := c.token.Get()
	if tokenStr != "" && !c.isTokenExpired() {
		metrics.RecordCacheHit("token")
		return tokenStr, nil
	}

	metrics.RecordCacheMiss("token")

	if !c.refreshing.CompareAndSwap(false, true) {
		c.mu.RLock()
		defer c.mu.RUnlock()
		tokenStr = c.token.Get()
		if tokenStr != "" && !c.isTokenExpired() {
			return tokenStr, nil
		}
		return "", fmt.Errorf("token 刷新进行中，请稍后重试")
	}
	defer c.refreshing.Store(false)

	start := time.Now()
	tokenStr, err := c.doRefreshToken()
	if err != nil {
		c.connected.Store(false)
		logger.Error("Failed to get token after retries", zap.Error(err))
		return "", err
	}

	c.connected.Store(true)
	metrics.RecordCacheOperation("token", "refresh", time.Since(start))
	return tokenStr, nil
}

// RefreshToken 强制刷新 Token
func (c *Client) RefreshToken() (string, error) {
	if !c.refreshing.CompareAndSwap(false, true) {
		return "", fmt.Errorf("token 刷新进行中")
	}
	defer c.refreshing.Store(false)

	start := time.Now()
	tokenStr, err := c.doRefreshToken()
	if err != nil {
		c.connected.Store(false)
		return "", err
	}

	c.connected.Store(true)
	metrics.RecordCacheOperation("token", "refresh", time.Since(start))
	return tokenStr, nil
}

// doRefreshToken 在无锁状态下执行网络请求，完成后用短锁写入状态
func (c *Client) doRefreshToken() (string, error) {
	c.mu.RLock()
	account := c.account
	passwordStr := c.password.Get()
	sdk := c.sdkClient
	c.mu.RUnlock()

	start := time.Now()
	var token string
	var err error
	for i := 0; i < 3; i++ {
		token, err = sdk.GetToken(account, passwordStr)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	passwordStr = ""

	if err != nil {
		return "", err
	}

	c.mu.Lock()
	c.token.Set(token)
	c.tokenExpiry.Store(time.Now().Add(tokenTTL).Unix())
	c.sdkClient.SetToken(token)
	c.mu.Unlock()

	metrics.RecordTokenRefresh()
	logger.Info("Token refreshed successfully",
		zap.Duration("duration", time.Since(start)),
	)

	return token, nil
}

// UpdateConfig 更新客户端配置并异步刷新Token
func (c *Client) UpdateConfig(server, account, password string) error {
	if !strings.HasPrefix(server, "http://") && !strings.HasPrefix(server, "https://") {
		server = "http://" + server
	}
	if strings.HasSuffix(server, "/") {
		server = strings.TrimSuffix(server, "/")
	}

	c.mu.Lock()
	c.server = server
	c.account = account
	c.password.Set(password)
	c.sdkClient = zentao.NewClient(server)
	c.sdkClient.SetTimeout(120 * time.Second)
	c.token.Set("")
	c.tokenExpiry.Store(0)
	GlobalCache.Clear()
	c.mu.Unlock()

	go func() {
		if _, err := c.RefreshToken(); err != nil {
			logger.Warn("异步刷新Token失败", zap.Error(err))
		}
	}()

	return nil
}

func (c *Client) GetServer() string {
	return c.server
}

func (c *Client) IsConnected() bool {
	return c.connected.Load()
}

func (c *Client) withTokenRetry(operation string, call func(*zentao.Client) error) error {
	if _, err := c.getToken(); err != nil {
		return err
	}

	c.mu.RLock()
	sdk := c.sdkClient
	c.mu.RUnlock()

	err := call(sdk)
	if err == nil || !isAuthError(err) {
		return err
	}

	logger.Warn("Zentao token rejected, refreshing and retrying",
		zap.String("operation", operation),
		zap.Error(err),
	)

	if _, refreshErr := c.RefreshToken(); refreshErr != nil {
		return fmt.Errorf("%s失败，刷新Token失败: %w，原始错误: %v", operation, refreshErr, err)
	}

	c.mu.RLock()
	sdk = c.sdkClient
	c.mu.RUnlock()
	return call(sdk)
}

func (c *Client) withTokenRetryContext(ctx context.Context, operation string, call func(*zentao.Client) error) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if _, err := c.getToken(); err != nil {
		return err
	}

	c.mu.RLock()
	sdk := c.sdkClient
	c.mu.RUnlock()

	err := call(sdk)
	if err == nil || !isAuthError(err) {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	logger.Warn("Zentao token rejected, refreshing and retrying",
		zap.String("operation", operation),
		zap.Error(err),
	)

	if _, refreshErr := c.RefreshToken(); refreshErr != nil {
		return fmt.Errorf("%s失败，刷新Token失败: %w，原始错误: %v", operation, refreshErr, err)
	}

	c.mu.RLock()
	sdk = c.sdkClient
	c.mu.RUnlock()
	return call(sdk)
}

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	// 优先使用 SDK 提供的类型检查
	if zentao.IsAuthError(err) {
		return true
	}
	// 兜底：匹配错误信息中的认证相关关键字
	msg := strings.ToLower(err.Error())
	authMarkers := []string{
		"状态码: 401",
		"status code: 401",
		"unauthorized",
		"状态码: 403",
		"status code: 403",
		"forbidden",
		"authentication required",
	}
	for _, marker := range authMarkers {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}

// GetAccount 获取当前登录用户的账号
func (c *Client) GetAccount() string {
	return c.account
}

// GetProducts 获取产品列表
func (c *Client) GetProducts() ([]zentao.Product, error) {
	cacheKey := "zentao:products:all"

	result, err := GlobalCache.GetOrLoadWithLock(cacheKey, func() (interface{}, error) {
		metrics.RecordCacheMiss("products")
		start := time.Now()

		if _, err := c.getToken(); err != nil {
			return nil, err
		}

		// 先获取第一页，获取总数
		var firstPageResponse *zentao.ProductListResponse
		if err := c.withTokenRetry("GetProducts", func(client *zentao.Client) error {
			var err error
			firstPageResponse, err = client.GetProducts(1, 100)
			return err
		}); err != nil {
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
				var pageResponse *zentao.ProductListResponse
				if err := c.withTokenRetry("GetProducts", func(client *zentao.Client) error {
					var err error
					pageResponse, err = client.GetProducts(page, pageSize)
					return err
				}); err != nil {
					metrics.RecordZentaoAPIRequest("products", "GET", time.Since(start), err)
					return nil, err
				}
				allProducts = append(allProducts, pageResponse.Products...)
			}
		}

		duration := time.Since(start)
		metrics.RecordCacheOperation("products", "fetch", duration)
		metrics.RecordZentaoAPIRequest("products", "GET", duration, nil)

		logger.Info("Products fetched",
			zap.Int("count", len(allProducts)),
			zap.Duration("duration", duration),
		)

		return allProducts, nil
	}, 5*time.Minute)

	if err != nil {
		return nil, err
	}
	metrics.RecordCacheHit("products")
	return result.([]zentao.Product), nil
}

// GetProduct 获取产品详情
func (c *Client) GetProduct(productID int) (*zentao.Product, error) {
	var result *zentao.Product
	err := c.withTokenRetry("GetProduct", func(client *zentao.Client) error {
		var err error
		result, err = client.GetProduct(productID)
		return err
	})
	return result, err
}

// GetAllProjects 获取所有项目列表
func (c *Client) GetAllProjects(limit int) ([]zentao.Project, error) {
	var response *zentao.ProjectListResponse
	err := c.withTokenRetry("GetAllProjects", func(client *zentao.Client) error {
		var err error
		response, err = client.GetAllProjects(1, limit)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Projects, nil
}

// GetProjectsByProduct 获取产品关联的项目列表
func (c *Client) GetProjectsByProduct(productID int, page, pageSize int) ([]zentao.Project, error) {
	cacheKey := DefaultKeyBuilder.Build("zentao:projects", strconv.Itoa(productID), strconv.Itoa(page), strconv.Itoa(pageSize))

	result, err := GlobalCache.GetOrLoadWithLock(cacheKey, func() (interface{}, error) {
		var response *zentao.ProjectListResponse
		err := c.withTokenRetry("GetProjectsByProduct", func(client *zentao.Client) error {
			var err error
			response, err = client.GetProjectsByProduct(productID, page, pageSize)
			return err
		})
		if err != nil {
			return nil, err
		}
		return response.Projects, nil
	}, 3*time.Minute)

	if err != nil {
		return nil, err
	}
	return result.([]zentao.Project), nil
}

// GetProject 获取项目详情
func (c *Client) GetProject(projectID int) (*zentao.Project, error) {
	var result *zentao.Project
	err := c.withTokenRetry("GetProject", func(client *zentao.Client) error {
		var err error
		result, err = client.GetProject(projectID)
		return err
	})
	return result, err
}

// GetBugs 获取产品的 Bug 列表
func (c *Client) GetBugs(productID int, page, pageSize int) ([]zentao.Bug, error) {
	cacheKey := DefaultKeyBuilder.Build("zentao:bugs", strconv.Itoa(productID), strconv.Itoa(page), strconv.Itoa(pageSize))

	result, err := GlobalCache.GetOrLoadWithLock(cacheKey, func() (interface{}, error) {
		var response *zentao.BugListResponse
		err := c.withTokenRetry("GetBugs", func(client *zentao.Client) error {
			var err error
			response, err = client.GetBugs(productID, page, pageSize)
			return err
		})
		if err != nil {
			return nil, err
		}
		return response.Bugs, nil
	}, 2*time.Minute)

	if err != nil {
		return nil, err
	}
	return result.([]zentao.Bug), nil
}

// GetBugsByProject 根据项目 ID 过滤 Bug 列表
func (c *Client) GetBugsByProject(productID, projectID int, page, pageSize int) ([]zentao.Bug, error) {
	var response *zentao.BugListResponse
	err := c.withTokenRetry("GetBugsByProject", func(client *zentao.Client) error {
		var err error
		response, err = client.GetBugsByProject(productID, projectID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Bugs, nil
}

// GetBugsByStatus 根据状态过滤 Bug 列表
func (c *Client) GetBugsByStatus(productID int, status string, page, pageSize int) ([]zentao.Bug, error) {
	var response *zentao.BugListResponse
	err := c.withTokenRetry("GetBugsByStatus", func(client *zentao.Client) error {
		var err error
		response, err = client.GetBugsByStatus(productID, status, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Bugs, nil
}

// SearchBugs 搜索 Bug（支持多条件过滤）
func (c *Client) SearchBugs(params zentao.BugSearchParams) ([]zentao.Bug, error) {
	var response *zentao.BugListResponse
	err := c.withTokenRetry("SearchBugs", func(client *zentao.Client) error {
		var err error
		response, err = client.SearchBugs(params)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Bugs, nil
}

// GetBug 获取 Bug 详情
func (c *Client) GetBug(bugID int) (*zentao.Bug, error) {
	var result *zentao.Bug
	err := c.withTokenRetry("GetBug", func(client *zentao.Client) error {
		var err error
		result, err = client.GetBug(bugID)
		return err
	})
	return result, err
}

// GetStoriesByProduct 获取产品的需求列表
func (c *Client) GetStoriesByProduct(productID int, page, pageSize int) ([]zentao.Story, error) {
	cacheKey := DefaultKeyBuilder.Build("zentao:stories", strconv.Itoa(productID), strconv.Itoa(page), strconv.Itoa(pageSize))

	result, err := GlobalCache.GetOrLoadWithLock(cacheKey, func() (interface{}, error) {
		var response *zentao.StoryListResponse
		err := c.withTokenRetry("GetStoriesByProduct", func(client *zentao.Client) error {
			var err error
			response, err = client.GetStoriesByProduct(productID, page, pageSize)
			return err
		})
		if err != nil {
			return nil, err
		}
		return response.Stories, nil
	}, 2*time.Minute)

	if err != nil {
		return nil, err
	}
	return result.([]zentao.Story), nil
}

// GetStoriesByProject 获取项目的需求列表
func (c *Client) GetStoriesByProject(projectID int, page, pageSize int) ([]zentao.Story, error) {
	var response *zentao.StoryListResponse
	err := c.withTokenRetry("GetStoriesByProject", func(client *zentao.Client) error {
		var err error
		response, err = client.GetStoriesByProject(projectID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Stories, nil
}

// GetStoriesByExecution 获取执行的需求列表
func (c *Client) GetStoriesByExecution(executionID int, page, pageSize int) ([]zentao.Story, error) {
	var response *zentao.StoryListResponse
	err := c.withTokenRetry("GetStoriesByExecution", func(client *zentao.Client) error {
		var err error
		response, err = client.GetStoriesByExecution(executionID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Stories, nil
}

// GetStory 获取需求详情
func (c *Client) GetStory(storyID int) (*zentao.Story, error) {
	var result *zentao.Story
	err := c.withTokenRetry("GetStory", func(client *zentao.Client) error {
		var err error
		result, err = client.GetStory(storyID)
		return err
	})
	return result, err
}

// GetTasks 获取执行的任务列表
func (c *Client) GetTasks(executionID int, page, pageSize int) ([]zentao.Task, error) {
	cacheKey := DefaultKeyBuilder.Build("zentao:tasks", strconv.Itoa(executionID), strconv.Itoa(page), strconv.Itoa(pageSize))

	result, err := GlobalCache.GetOrLoadWithLock(cacheKey, func() (interface{}, error) {
		var response *zentao.TaskListResponse
		err := c.withTokenRetry("GetTasks", func(client *zentao.Client) error {
			var err error
			response, err = client.GetTasks(executionID, page, pageSize)
			return err
		})
		if err != nil {
			return nil, err
		}
		return response.Tasks, nil
	}, 2*time.Minute)

	if err != nil {
		return nil, err
	}
	return result.([]zentao.Task), nil
}

// GetTask 获取任务详情
func (c *Client) GetTask(taskID int) (*zentao.Task, error) {
	var result *zentao.Task
	err := c.withTokenRetry("GetTask", func(client *zentao.Client) error {
		var err error
		result, err = client.GetTask(taskID)
		return err
	})
	return result, err
}

// GetExecutions 获取执行列表
func (c *Client) GetExecutions(projectID int, page, pageSize int) ([]zentao.Execution, error) {
	cacheKey := DefaultKeyBuilder.Build("zentao:executions", strconv.Itoa(projectID), strconv.Itoa(page), strconv.Itoa(pageSize))

	result, err := GlobalCache.GetOrLoadWithLock(cacheKey, func() (interface{}, error) {
		var response *zentao.ExecutionListResponse
		err := c.withTokenRetry("GetExecutions", func(client *zentao.Client) error {
			var err error
			response, err = client.GetExecutions(projectID, page, pageSize)
			return err
		})
		if err != nil {
			return nil, err
		}
		return response.Executions, nil
	}, 3*time.Minute)

	if err != nil {
		return nil, err
	}
	return result.([]zentao.Execution), nil
}

// GetUsers 获取用户列表（支持分页）
func (c *Client) GetUsers(page, pageSize int) (*zentao.UserListResponse, error) {
	// 先获取全部用户（带缓存）
	allUsers, err := c.GetUsersAll()
	if err != nil {
		return nil, err
	}

	metrics.RecordCacheHit("users")

	// 计算分页
	total := len(allUsers)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		return &zentao.UserListResponse{
			Users: []zentao.User{},
			Page:  page,
			Total: total,
			Limit: pageSize,
		}, nil
	}
	if end > total {
		end = total
	}

	return &zentao.UserListResponse{
		Users: allUsers[start:end],
		Page:  page,
		Total: total,
		Limit: pageSize,
	}, nil
}

// GetUsersAll 获取所有用户列表
func (c *Client) GetUsersAll() ([]zentao.User, error) {
	cacheKey := "zentao:users:all"

	result, err := GlobalCache.GetOrLoadWithLock(cacheKey, func() (interface{}, error) {
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
			var response *zentao.UserListResponse
			if err := c.withTokenRetry("GetUsers", func(client *zentao.Client) error {
				var err error
				response, err = client.GetUsers(currentPage, pageSize)
				return err
			}); err != nil {
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

		duration := time.Since(startTime)
		metrics.RecordCacheOperation("users", "fetch", duration)
		metrics.RecordZentaoAPIRequest("users", "GET", duration, nil)

		logger.Info("Users fetched",
			zap.Int("count", len(allUsers)),
			zap.Duration("duration", duration),
		)

		return allUsers, nil
	}, 10*time.Minute)

	if err != nil {
		return nil, err
	}
	return result.([]zentao.User), nil
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
				if consumed, ok := t.Consumed.(float64); ok && consumed > 0 {
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

// ========== Dashboard 聚合查询方法 ==========

// ExecutionContext 执行上下文（包含执行信息和关联的项目）
type ExecutionContext struct {
	Exec     zentao.Execution
	ProjName string
	ExecName string
}

func getTaskFloat64(v interface{}) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

// GetAllBugs 获取产品全部 Bug（自动翻页）
func (c *Client) GetAllBugs(productID int) ([]zentao.Bug, error) {
	var all []zentao.Bug
	page := 1
	for {
		bugs, err := c.GetBugs(productID, page, 100)
		if err != nil {
			return all, err
		}
		all = append(all, bugs...)
		if len(bugs) < 100 {
			break
		}
		page++
	}
	return all, nil
}

// GetAllBugsByProject 获取项目全部 Bug
func (c *Client) GetAllBugsByProject(projectID int) ([]zentao.Bug, error) {
	return c.GetAllBugsByProjectWithProduct(0, projectID)
}

// GetAllBugsByProjectWithProduct 获取项目全部 Bug（指定产品ID）
func (c *Client) GetAllBugsByProjectWithProduct(productID int, projectID int) ([]zentao.Bug, error) {
	var all []zentao.Bug
	page := 1
	for {
		bugs, err := c.GetBugsByProject(productID, projectID, page, 100)
		if err != nil {
			return all, err
		}
		all = append(all, bugs...)
		if len(bugs) < 100 {
			break
		}
		page++
	}
	return all, nil
}

// GetAllStories 获取产品全部需求（自动翻页）
func (c *Client) GetAllStories(productID int) ([]zentao.Story, error) {
	var all []zentao.Story
	page := 1
	for {
		stories, err := c.GetStoriesByProduct(productID, page, 100)
		if err != nil {
			return all, err
		}
		all = append(all, stories...)
		if len(stories) < 100 {
			break
		}
		page++
	}
	return all, nil
}

// GetAllTasksByProject 获取项目全部任务（自动翻页）
func (c *Client) GetAllTasksByProject(projectID int) ([]zentao.Task, error) {
	var all []zentao.Task
	executions, err := c.GetExecutions(projectID, 1, 1000)
	if err != nil {
		return nil, err
	}
	for _, execution := range executions {
		tasks, err := c.GetTasks(execution.ID, 1, 10000)
		if err != nil {
			continue
		}
		all = append(all, tasks...)
	}
	return all, nil
}

// GetAllTasksByProduct 获取产品全部任务（自动翻页）
func (c *Client) GetAllTasksByProduct(productID int) ([]zentao.Task, error) {
	var all []zentao.Task
	projects, err := c.GetProjectsByProduct(productID, 1, 2000)
	if err != nil {
		return nil, err
	}
	for _, project := range projects {
		tasks, err := c.GetAllTasksByProject(project.ID)
		if err != nil {
			continue
		}
		all = append(all, tasks...)
	}
	return all, nil
}

// GetAllStoriesByProject 获取项目全部需求（自动翻页）
func (c *Client) GetAllStoriesByProject(projectID int) ([]zentao.Story, error) {
	var all []zentao.Story
	executions, err := c.GetExecutions(projectID, 1, 1000)
	if err != nil {
		return nil, err
	}
	for _, execution := range executions {
		stories, err := c.GetStoriesByExecution(execution.ID, 1, 10000)
		if err != nil {
			continue
		}
		all = append(all, stories...)
	}
	return all, nil
}

// GetExecutionsByProduct 按产品获取所有执行
func (c *Client) GetExecutionsByProduct(productID int) ([]ExecutionContext, error) {
	var result []ExecutionContext
	projects, err := c.GetProjectsByProduct(productID, 1, 200)
	if err != nil {
		return nil, err
	}
	for _, p := range projects {
		executions, err := c.GetExecutions(p.ID, 1, 200)
		if err != nil {
			continue
		}
		for _, e := range executions {
			result = append(result, ExecutionContext{
				Exec:     e,
				ProjName: p.Name,
				ExecName: e.Name,
			})
		}
	}
	return result, nil
}

// GetTaskEfforts 获取任务的工时记录
func (c *Client) GetTaskEfforts(taskID int) ([]zentao.EffortEntry, error) {
	var result []zentao.EffortEntry
	err := c.withTokenRetry("GetTaskEfforts", func(client *zentao.Client) error {
		var err error
		result, err = client.GetTaskEfforts(taskID)
		return err
	})
	return result, err
}

// ========== Context-aware 方法（用于 Dashboard 等需要取消的场景） ==========

// GetProductContext 获取产品详情（支持 context 取消）
func (c *Client) GetProductContext(ctx context.Context, productID int) (*zentao.Product, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var result *zentao.Product
	err := c.withTokenRetryContext(ctx, "GetProduct", func(client *zentao.Client) error {
		var err error
		result, err = client.GetProduct(productID)
		return err
	})
	return result, err
}

// GetProjectContext 获取项目详情（支持 context 取消）
func (c *Client) GetProjectContext(ctx context.Context, projectID int) (*zentao.Project, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var result *zentao.Project
	err := c.withTokenRetryContext(ctx, "GetProject", func(client *zentao.Client) error {
		var err error
		result, err = client.GetProject(projectID)
		return err
	})
	return result, err
}

// GetProjectsByProductContext 获取产品关联的项目（支持 context 取消）
func (c *Client) GetProjectsByProductContext(ctx context.Context, productID int, page, pageSize int) ([]zentao.Project, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var response *zentao.ProjectListResponse
	err := c.withTokenRetryContext(ctx, "GetProjectsByProduct", func(client *zentao.Client) error {
		var err error
		response, err = client.GetProjectsByProduct(productID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Projects, nil
}

// GetExecutionsContext 获取执行列表（支持 context 取消）
func (c *Client) GetExecutionsContext(ctx context.Context, projectID int, page, pageSize int) ([]zentao.Execution, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var response *zentao.ExecutionListResponse
	err := c.withTokenRetryContext(ctx, "GetExecutions", func(client *zentao.Client) error {
		var err error
		response, err = client.GetExecutions(projectID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Executions, nil
}

// GetTasksContext 获取执行的任务列表（支持 context 取消）
func (c *Client) GetTasksContext(ctx context.Context, executionID int, page, pageSize int) ([]zentao.Task, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var response *zentao.TaskListResponse
	err := c.withTokenRetryContext(ctx, "GetTasks", func(client *zentao.Client) error {
		var err error
		response, err = client.GetTasks(executionID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Tasks, nil
}

// GetBugsContext 获取 Bug 列表（支持 context 取消）
func (c *Client) GetBugsContext(ctx context.Context, productID int, page, pageSize int) ([]zentao.Bug, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var response *zentao.BugListResponse
	err := c.withTokenRetryContext(ctx, "GetBugs", func(client *zentao.Client) error {
		var err error
		response, err = client.GetBugs(productID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Bugs, nil
}

// GetStoriesByProductContext 获取产品的需求列表（支持 context 取消）
func (c *Client) GetStoriesByProductContext(ctx context.Context, productID int, page, pageSize int) ([]zentao.Story, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var response *zentao.StoryListResponse
	err := c.withTokenRetryContext(ctx, "GetStoriesByProduct", func(client *zentao.Client) error {
		var err error
		response, err = client.GetStoriesByProduct(productID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Stories, nil
}

// GetTaskEffortsContext 获取任务的工时记录（支持 context 取消）
func (c *Client) GetTaskEffortsContext(ctx context.Context, taskID int) ([]zentao.EffortEntry, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var result []zentao.EffortEntry
	err := c.withTokenRetryContext(ctx, "GetTaskEfforts", func(client *zentao.Client) error {
		var err error
		result, err = client.GetTaskEfforts(taskID)
		return err
	})
	return result, err
}

// GetAllBugsContext 获取产品全部 Bug（支持 context 取消，自动翻页）
func (c *Client) GetAllBugsContext(ctx context.Context, productID int) ([]zentao.Bug, error) {
	var all []zentao.Bug
	page := 1
	for {
		select {
		case <-ctx.Done():
			return all, ctx.Err()
		default:
		}
		bugs, err := c.GetBugsContext(ctx, productID, page, 100)
		if err != nil {
			return all, err
		}
		all = append(all, bugs...)
		if len(bugs) < 100 {
			break
		}
		page++
	}
	return all, nil
}

// GetAllBugsByProjectContext 获取项目全部 Bug（支持 context 取消）
func (c *Client) GetAllBugsByProjectContext(ctx context.Context, projectID int) ([]zentao.Bug, error) {
	var all []zentao.Bug
	page := 1
	for {
		select {
		case <-ctx.Done():
			return all, ctx.Err()
		default:
		}
		var response *zentao.BugListResponse
		err := c.withTokenRetryContext(ctx, "GetBugsByProject", func(client *zentao.Client) error {
			var err error
			response, err = client.GetBugsByProject(0, projectID, page, 100)
			return err
		})
		if err != nil {
			return all, err
		}
		all = append(all, response.Bugs...)
		if len(response.Bugs) < 100 {
			break
		}
		page++
	}
	return all, nil
}

// GetAllStoriesContext 获取产品全部需求（支持 context 取消，自动翻页）
func (c *Client) GetAllStoriesContext(ctx context.Context, productID int) ([]zentao.Story, error) {
	var all []zentao.Story
	page := 1
	for {
		select {
		case <-ctx.Done():
			return all, ctx.Err()
		default:
		}
		stories, err := c.GetStoriesByProductContext(ctx, productID, page, 100)
		if err != nil {
			return all, err
		}
		all = append(all, stories...)
		if len(stories) < 100 {
			break
		}
		page++
	}
	return all, nil
}

// GetExecutionsByProductContext 按产品获取所有执行（支持 context 取消）
func (c *Client) GetExecutionsByProductContext(ctx context.Context, productID int) ([]ExecutionContext, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var result []ExecutionContext
	projects, err := c.GetProjectsByProductContext(ctx, productID, 1, 200)
	if err != nil {
		return nil, err
	}
	for _, p := range projects {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}
		executions, err := c.GetExecutionsContext(ctx, p.ID, 1, 200)
		if err != nil {
			continue
		}
		for _, e := range executions {
			result = append(result, ExecutionContext{
				Exec:     e,
				ProjName: p.Name,
				ExecName: e.Name,
			})
		}
	}
	return result, nil
}
