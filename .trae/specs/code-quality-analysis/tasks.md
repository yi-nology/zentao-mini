# Tasks

- [x] Task 1: 架构重构 - 统一应用入口
  - [x] SubTask 1.1: 创建统一的应用管理器（Application Manager）
  - [x] SubTask 1.2: 抽象Wails和HTTP Server的启动逻辑
  - [x] SubTask 1.3: 实现优雅关闭机制

- [x] Task 2: 引入依赖注入
  - [x] SubTask 2.1: 选择并集成依赖注入框架（如wire或dig）
  - [x] SubTask 2.2: 定义依赖注入容器
  - [x] SubTask 2.3: 重构handler初始化代码

- [x] Task 3: 消除代码重复
  - [x] SubTask 3.1: 创建分页中间件/工具函数
  - [x] SubTask 3.2: 提取通用的数据转换逻辑
  - [x] SubTask 3.3: 创建通用的过滤和排序工具

- [x] Task 4: 统一错误处理
  - [x] SubTask 4.1: 定义统一的错误码规范
  - [x] SubTask 4.2: 创建错误处理中间件
  - [x] SubTask 4.3: 实现错误日志记录和敏感信息过滤

- [x] Task 5: 安全加固
  - [x] SubTask 5.1: 移除硬编码密钥，使用环境变量
  - [x] SubTask 5.2: 配置严格的CORS策略
  - [x] SubTask 5.3: 实现敏感信息的安全存储
  - [x] SubTask 5.4: 添加请求限流中间件

- [x] Task 6: 性能优化
  - [x] SubTask 6.1: 优化API调用，支持服务端过滤
  - [x] SubTask 6.2: 实现优雅的并发控制（worker pool）
  - [x] SubTask 6.3: 添加缓存机制

- [x] Task 7: API规范化
  - [x] SubTask 7.1: 统一响应格式
  - [x] SubTask 7.2: 统一参数命名风格
  - [x] SubTask 7.3: 引入API版本控制

- [x] Task 8: 代码组织优化
  - [x] SubTask 8.1: 分离请求参数模型（DTO）和响应模型（VO）
  - [x] SubTask 8.2: 引入Service层，分离业务逻辑
  - [x] SubTask 8.3: 重组包结构

- [x] Task 9: 配置管理改进
  - [x] SubTask 9.1: 引入配置管理库（如viper）
  - [x] SubTask 9.2: 统一配置来源和命名规范
  - [x] SubTask 9.3: 添加配置验证

- [x] Task 10: 日志和监控
  - [x] SubTask 10.1: 引入结构化日志库（如zap或logrus）
  - [x] SubTask 10.2: 添加请求追踪ID
  - [x] SubTask 10.3: 实现性能指标收集

- [x] Task 11: 测试覆盖
  - [x] SubTask 11.1: 为核心业务逻辑编写单元测试
  - [x] SubTask 11.2: 为handler编写集成测试
  - [x] SubTask 11.3: 设置测试覆盖率目标（70%+）

- [x] Task 12: 代码质量改进
  - [x] SubTask 12.1: 消除魔法数字，定义常量
  - [x] SubTask 12.2: 添加必要的代码注释
  - [x] SubTask 12.3: 配置代码检查工具（golangci-lint）

# Task Dependencies
- [Task 2] depends on [Task 1] - 依赖注入需要在统一架构后实施
- [Task 3] depends on [Task 2] - 消除重复需要依赖注入支持
- [Task 4] depends on [Task 2] - 错误处理中间件需要依赖注入
- [Task 6] depends on [Task 3] - 性能优化需要先消除重复代码
- [Task 7] depends on [Task 4] - API规范化需要统一错误处理
- [Task 8] depends on [Task 2] - 代码组织优化需要依赖注入
- [Task 11] depends on [Task 8] - 测试需要在代码组织优化后编写
