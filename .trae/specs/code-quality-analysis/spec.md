# 代码质量分析和重构建议 Spec

## Why
当前代码库存在多处架构设计、代码质量、安全性和性能方面的问题，需要进行系统性分析和重构建议，以提高代码的可维护性、安全性和性能。

## What Changes
- 识别并记录所有代码不合理之处
- 提供详细的重构建议
- 建立代码质量改进计划

## Impact
- Affected specs: 整个后端架构
- Affected code: backend/core目录下所有文件

## ADDED Requirements

### Requirement: 架构设计问题
系统应遵循清晰的分层架构，避免职责混乱。

#### Scenario: 应用模式混乱
- **WHEN** 查看项目结构时
- **THEN** 发现同时存在Wails桌面应用模式和HTTP服务器模式，导致代码职责不清

**问题详情**:
1. `main.go` 和 `backend/cmd/server/main.go` 提供了两种不同的启动方式
2. Wails应用和HTTP服务器共享相同的handlers，但没有清晰的抽象层
3. 缺少统一的应用入口和生命周期管理

#### Scenario: 缺少依赖注入
- **WHEN** 查看handler初始化代码时
- **THEN** 发现所有依赖都是手动创建，没有使用依赖注入容器

**问题详情**:
1. `backend/cmd/server/main.go` 中手动创建了7个handler实例
2. `routes.go` 中又重复创建了一次handler实例
3. 依赖关系不清晰，难以测试和替换实现

### Requirement: 代码重复问题
系统应避免代码重复，提取公共逻辑。

#### Scenario: 分页逻辑重复
- **WHEN** 查看多个handler的Get方法时
- **THEN** 发现分页参数解析和计算逻辑在多处重复

**问题详情**:
1. `bugs.go`、`tasks.go`、`stories.go` 中都有相同的分页参数解析代码（第54-63行）
2. 分页计算逻辑重复（计算start、end、边界检查）
3. 没有提取为公共函数或中间件

#### Scenario: 数据转换逻辑重复
- **WHEN** 查看模型转换代码时
- **THEN** 发现每个handler都有类似的转换逻辑

**问题详情**:
1. 每个handler都有convertXXX函数
2. 转换逻辑简单但重复
3. 可以使用泛型或反射简化

### Requirement: 错误处理一致性
系统应提供统一的错误处理机制。

#### Scenario: 错误响应格式不一致
- **WHEN** 查看不同handler的错误处理时
- **THEN** 发现错误响应格式不统一

**问题详情**:
1. `products.go` 使用 `models.Response` 结构
2. `timelog.go` 使用 `gin.H{"status": "error", "error": "..."}`
3. 缺少统一的错误码和错误消息规范

#### Scenario: 错误信息泄露
- **WHEN** 发生内部错误时
- **THEN** 直接返回底层错误信息给客户端

**问题详情**:
1. `products.go:36` 直接返回 `err.Error()`
2. 可能泄露敏感信息（如数据库连接字符串）
3. 应该返回友好的错误消息，记录详细错误到日志

### Requirement: 安全性问题
系统应遵循安全最佳实践。

#### Scenario: 硬编码加密密钥
- **WHEN** 查看 `initialization/init.go` 时
- **THEN** 发现加密密钥硬编码在代码中

**问题详情**:
1. `init.go:56` 硬编码密钥 `"Zhangyi@Kylin999-"`
2. 密钥应该从环境变量或配置文件读取
3. 代码库泄露会导致所有加密数据被破解

#### Scenario: CORS配置过于宽松
- **WHEN** 查看 `routes.go` 时
- **THEN** 发现CORS允许所有来源

**问题详情**:
1. `routes.go:29` 设置 `AllowOrigins: []string{"*"}`
2. 生产环境应该限制允许的域名
3. 结合 `AllowCredentials: true` 存在安全风险

#### Scenario: 敏感信息内存存储
- **WHEN** 查看Client结构时
- **THEN** 发现密码明文存储在内存中

**问题详情**:
1. `client.go:17` 存储明文密码
2. 内存转储可能泄露密码
3. 应该在使用后立即清除或使用更安全的方式

### Requirement: 性能优化
系统应优化性能瓶颈。

#### Scenario: 客户端过滤导致性能问题
- **WHEN** 查询大量数据时
- **THEN** 先获取全部数据再在客户端过滤

**问题详情**:
1. `bugs.go:84` 获取1000条数据用于筛选
2. `tasks.go:95` 获取500条任务
3. `stories.go:78` 获取500条需求
4. 应该在API层面支持过滤参数

#### Scenario: 缺少请求限流
- **WHEN** 查看路由配置时
- **THEN** 发现没有限流中间件

**问题详情**:
1. 恶意用户可以发起大量请求
2. 可能导致禅道API被封禁
3. 应该添加基于IP或用户的限流

#### Scenario: 并发控制不优雅
- **WHEN** 查看 `GetTimelogAnalysis` 时
- **THEN** 发现使用channel限制并发数

**问题详情**:
1. `client.go:646-717` 使用channel手动控制并发
2. Go有更优雅的并发控制方式（如worker pool）
3. 错误处理不完善，goroutine泄漏风险

### Requirement: API设计规范
系统应遵循RESTful API设计规范。

#### Scenario: 响应格式不统一
- **WHEN** 查看不同API的响应时
- **THEN** 发现响应结构不一致

**问题详情**:
1. 有的返回 `models.Response`
2. 有的返回 `gin.H`
3. 缺少统一的响应包装器

#### Scenario: 参数命名不一致
- **WHEN** 查看API参数时
- **THEN** 发现参数命名风格混乱

**问题详情**:
1. `productId` vs `productID`（驼峰命名不一致）
2. `assignedTo` vs `assigned_to`（驼峰vs下划线）
3. 应该统一使用一种命名风格

#### Scenario: 缺少API版本控制
- **WHEN** 查看路由定义时
- **THEN** 发现所有API都在 `/api` 下

**问题详情**:
1. 没有版本号（如 `/api/v1`）
2. API变更会破坏向后兼容性
3. 应该引入版本控制机制

### Requirement: 代码组织规范
系统应有清晰的代码组织结构。

#### Scenario: 模型职责不清
- **WHEN** 查看 `models/models.go` 时
- **THEN** 发现请求参数和响应模型混在一起

**问题详情**:
1. `BugSearchParams` 是请求参数，不应该在models包
2. 缺少DTO（Data Transfer Object）和VO（Value Object）分离
3. 业务模型和API模型耦合

#### Scenario: 业务逻辑混入Handler
- **WHEN** 查看handler代码时
- **THEN** 发现业务逻辑直接写在handler中

**问题详情**:
1. `bugs.go:67-134` 包含大量业务逻辑
2. 数据过滤、分页逻辑应该在service层
3. Handler应该只负责HTTP请求/响应处理

### Requirement: 配置管理规范
系统应有统一的配置管理。

#### Scenario: 配置分散
- **WHEN** 查看配置使用时
- **THEN** 发现配置来源分散

**问题详情**:
1. 环境变量、配置文件、硬编码混用
2. 缺少配置验证
3. 没有配置热更新机制

#### Scenario: 环境变量命名不规范
- **WHEN** 查看环境变量时
- **THEN** 发现命名风格不一致

**问题详情**:
1. `ZENTAO_SERVER`、`AUTH_CONFIG_PATH`、`PORT`
2. 缺少统一前缀
3. 应该使用如 `APP_` 或 `ZENTAO_MINI_` 前缀

### Requirement: 日志和监控
系统应有完善的日志和监控。

#### Scenario: 缺少结构化日志
- **WHEN** 查看日志输出时
- **THEN** 发现使用简单的log.Printf

**问题详情**:
1. 没有日志级别（DEBUG、INFO、ERROR）
2. 没有结构化日志（JSON格式）
3. 缺少请求追踪ID

#### Scenario: 缺少性能监控
- **WHEN** 查看代码时
- **THEN** 发现没有性能指标收集

**问题详情**:
1. 没有请求耗时统计
2. 没有数据库/API调用监控
3. 缺少健康检查详情

### Requirement: 测试覆盖
系统应有充分的测试覆盖。

#### Scenario: 缺少单元测试
- **WHEN** 查看测试文件时
- **THEN** 发现几乎没有测试代码

**问题详情**:
1. 只有 `test_decrypt.go` 一个测试文件
2. 核心业务逻辑没有测试
3. 应该至少有70%以上的测试覆盖率

### Requirement: 魔法数字问题
系统应避免使用魔法数字。

#### Scenario: 硬编码的数值
- **WHEN** 查看代码时
- **THEN** 发现大量硬编码的数字

**问题详情**:
1. `bugs.go:84` limit=1000
2. `tasks.go:95` limit=500
3. `client.go:40` timeout=120秒
4. 应该定义为常量或配置项

### Requirement: 注释和文档
系统应有充分的注释和文档。

#### Scenario: 缺少代码注释
- **WHEN** 查看复杂逻辑时
- **THEN** 发现缺少解释性注释

**问题详情**:
1. `GetTimelogAnalysis` 函数有200多行，但注释很少
2. 复杂的并发逻辑没有说明
3. 应该添加必要的注释说明设计意图
