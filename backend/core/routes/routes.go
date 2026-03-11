package routes

import (
	"bufio"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"chandao-mini/backend/core/handlers"
	"chandao-mini/backend/core/initialization"
	"chandao-mini/backend/core/zentao"
)

// SetupRouter configures the router with all routes
func SetupRouter(initService *initialization.InitService, zentaoClient *zentao.Client) *gin.Engine {
	// 设置 Gin 模式
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)

	// 创建 Gin 引擎
	r := gin.Default()

	// 配置 CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "Token"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = true
	r.Use(cors.New(corsConfig))

	// 初始化处理器
	productHandler := handlers.NewProductHandler(zentaoClient)
	projectHandler := handlers.NewProjectHandler(zentaoClient)
	executionHandler := handlers.NewExecutionHandler(zentaoClient)
	bugHandler := handlers.NewBugHandler(zentaoClient)
	storyHandler := handlers.NewStoryHandler(zentaoClient)
	taskHandler := handlers.NewTaskHandler(zentaoClient)
	userHandler := handlers.NewUserHandler(zentaoClient)
	timelogHandler := handlers.NewTimelogHandler(zentaoClient)

	// 健康检查接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "chandao-mini backend is running",
		})
	})

	// API 路由组
	api := r.Group("/api")
	{
		// 初始化相关接口
		api.POST("/init/upload", func(c *gin.Context) {
			// 接收上传的文件
			file, err := c.FormFile("configFile")
			if err != nil {
				c.JSON(400, gin.H{
					"code":    400,
					"message": "请选择要上传的文件",
					"data":    nil,
				})
				return
			}

			// 直接读取上传文件流，无需落盘
			fileData := make([]byte, file.Size)
			f, err := file.Open()
			if err != nil {
				c.JSON(500, gin.H{
					"code":    500,
					"message": "打开上传文件失败",
					"data":    nil,
				})
				return
			}
			defer f.Close()
			fileDataReader := bufio.NewReader(f)
			_, err = fileDataReader.Read(fileData)
			if err != nil {

			}

			// 重新加载配置
			authConfig, err := initService.LoadEncryptedConfig(fileData)
			if err != nil {
				// 记录错误但不返回给客户端，因为响应已经发送
				c.JSON(500, gin.H{
					"code":    500,
					"message": "加载认证配置失败",
					"data":    nil,
				})
				return
			}

			// 存储到数据库
			err = initService.StoreAuthConfig(fileData)
			if err != nil {
				// 记录错误但不返回给客户端，因为响应已经发送
				c.JSON(500, gin.H{
					"code":    500,
					"message": "存储认证配置失败",
					"data":    nil,
				})
				return
			}

			// 更新禅道客户端配置并刷新Token
			err = zentaoClient.UpdateConfig(authConfig.Domain, authConfig.Username, authConfig.Password)
			if err != nil {
				// 记录错误但不返回给客户端，因为响应已经发送
				c.JSON(500, gin.H{
					"code":    500,
					"message": "更新禅道配置失败",
					"data":    nil,
				})
				return
			}

			// 立即返回成功响应，后续操作在后台执行
			c.JSON(200, gin.H{
				"code":    200,
				"message": "初始化成功",
				"data":    nil,
			})
		})

		// 初始化状态接口
		api.GET("/init/status", func(c *gin.Context) {
			// 检查是否已初始化
			isFirstStart, err := initService.IsFirstStart()
			if err != nil {
				c.JSON(500, gin.H{
					"code":    500,
					"message": "检查初始化状态失败",
					"data":    nil,
				})
				return
			}

			c.JSON(200, gin.H{
				"code":    200,
				"message": "获取初始化状态成功",
				"data": gin.H{
					"isFirstStart": isFirstStart,
				},
			})
		})

		// 产品相关接口
		api.GET("/products", productHandler.GetProducts)

		// 项目相关接口
		api.GET("/projects", projectHandler.GetProjects)

		// 执行/迭代相关接口
		api.GET("/executions", executionHandler.GetExecutions)

		// Bug相关接口
		api.GET("/bugs", bugHandler.GetBugs)

		// 需求相关接口
		api.GET("/stories", storyHandler.GetStories)

		// 任务相关接口
		api.GET("/tasks", taskHandler.GetTasks)

		// 用户相关接口
		api.GET("/users", userHandler.GetUsers)
		api.GET("/users/all", userHandler.GetUsersAll)
		api.GET("/users/current", userHandler.GetCurrentUser)

		// 工时统计相关接口
		api.GET("/timelog/analysis", timelogHandler.GetTimelogAnalysis)
		api.GET("/timelog/dashboard", timelogHandler.GetTimelogDashboard)
		api.GET("/timelog/efforts", timelogHandler.GetTimelogEfforts)
	}

	// 静态文件服务 - 提供前端资源
	r.Static("/assets", "./frontend/dist/assets")

	// 前端路由处理 - 所有非API请求都返回index.html
	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	return r
}
