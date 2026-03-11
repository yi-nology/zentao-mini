# Tasks

- [x] Task 1: 创建后端项目结构
  - [x] SubTask 1.1: 初始化 Go 模块并创建 main.go
  - [x] SubTask 1.2: 配置依赖（gin, zentao sdk）
  - [x] SubTask 1.3: 创建项目目录结构

- [x] Task 2: 实现后端禅道 SDK 集成
  - [x] SubTask 2.1: 创建 zentao client 初始化模块
  - [x] SubTask 2.2: 实现 Token 获取和缓存逻辑
  - [x] SubTask 2.3: 封装 SDK 调用方法

- [x] Task 3: 实现后端 API 接口
  - [x] SubTask 3.1: 实现 GET /api/products 接口
  - [x] SubTask 3.2: 实现 GET /api/projects 接口
  - [x] SubTask 3.3: 实现 GET /api/bugs 接口（支持筛选参数）
  - [x] SubTask 3.4: 实现 GET /api/stories 接口（支持筛选参数）
  - [x] SubTask 3.5: 实现 GET /api/tasks 接口（支持筛选参数）
  - [x] SubTask 3.6: 配置 CORS 跨域支持

- [x] Task 4: 创建前端 Vue3 项目
  - [x] SubTask 4.1: 使用 Vite 初始化 Vue3 项目
  - [x] SubTask 4.2: 安装依赖（vue-router, axios, element-plus）
  - [x] SubTask 4.3: 配置项目目录结构

- [x] Task 5: 实现前端页面和组件
  - [x] SubTask 5.1: 创建主布局和侧边栏导航
  - [x] SubTask 5.2: 创建 Bug 查询页面（含筛选器、表格）
  - [x] SubTask 5.3: 创建需求查询页面（含筛选器、表格）
  - [x] SubTask 5.4: 创建任务查询页面（含筛选器、表格）
  - [x] SubTask 5.5: 封装 API 请求模块

- [x] Task 6: 集成测试和优化
  - [x] SubTask 6.1: 联调前后端接口
  - [x] SubTask 6.2: 优化 UI 样式和交互
  - [x] SubTask 6.3: 添加错误处理和加载状态

# Task Dependencies

- Task 2 依赖 Task 1
- Task 3 依赖 Task 2
- Task 4 可并行于 Task 1-3
- Task 5 依赖 Task 4 和 Task 3
- Task 6 依赖 Task 3 和 Task 5
