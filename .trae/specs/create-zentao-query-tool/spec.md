# 禅道 Bug/需求查询工具 Spec

## Why

需要开发一个禅道 Bug 和需求查询工具，方便团队成员快速查询禅道系统中的任务状态、Bug 详情和需求信息。工具采用前后端分离架构，前端使用 Vue3 + Vite，后端使用 Golang + Gin 框架。

## What Changes

- 创建后端 Golang Gin 服务，集成禅道 SDK
- 创建前端 Vue3 + Vite 项目
- 实现 Bug 查询功能（按产品、项目、状态筛选）
- 实现需求查询功能（按产品、项目、执行筛选）
- 实现任务查询功能
- 实现用户认证和 Token 管理
- 实现响应式 UI 界面

## Impact

- 新增后端 API 服务
- 新增前端 Web 应用
- 依赖禅道 SDK: `github.com/yi-nology/common/biz/zentao`
- 禅道服务器: `https://pm.kylin.com`
- 认证信息: zhangyi01 / ZYchandao2025!

## ADDED Requirements

### Requirement: 后端 API 服务

The system SHALL provide a Golang Gin backend service that integrates with Zentao SDK.

#### Scenario: 认证管理
- **WHEN** 服务启动时
- **THEN** 使用硬编码的账号密码获取禅道 Token 并缓存

#### Scenario: Bug 查询 API
- **WHEN** 用户调用 GET /api/bugs
- **THEN** 返回 Bug 列表，支持按 productID、projectID、status 筛选

#### Scenario: 需求查询 API
- **WHEN** 用户调用 GET /api/stories
- **THEN** 返回需求列表，支持按 productID、projectID、executionID 筛选

#### Scenario: 任务查询 API
- **WHEN** 用户调用 GET /api/tasks
- **THEN** 返回任务列表，支持按 executionID 筛选

#### Scenario: 产品和项目列表 API
- **WHEN** 用户调用 GET /api/products
- **THEN** 返回产品列表
- **WHEN** 用户调用 GET /api/projects
- **THEN** 返回项目列表

### Requirement: 前端 Web 应用

The system SHALL provide a Vue3 + Vite frontend application for querying Zentao data.

#### Scenario: Bug 查询页面
- **WHEN** 用户访问 /bugs 页面
- **THEN** 显示 Bug 列表，支持按产品、项目、状态筛选
- **AND** 显示 Bug 详情（标题、状态、指派人、创建时间等）

#### Scenario: 需求查询页面
- **WHEN** 用户访问 /stories 页面
- **THEN** 显示需求列表，支持按产品、项目筛选
- **AND** 显示需求详情（标题、状态、优先级、阶段等）

#### Scenario: 任务查询页面
- **WHEN** 用户访问 /tasks 页面
- **THEN** 显示任务列表，支持按执行筛选
- **AND** 显示任务详情（标题、状态、工时、进度等）

#### Scenario: 导航菜单
- **WHEN** 用户访问任意页面
- **THEN** 显示侧边栏导航菜单，包含 Bug、需求、任务入口

### Requirement: 禅道 SDK 集成

The system SHALL use `github.com/yi-nology/common/biz/zentao` SDK for API calls.

#### Scenario: SDK 初始化
- **WHEN** 后端服务启动
- **THEN** 初始化 Zentao Client，配置服务器地址 `https://pm.kylin.com`
- **AND** 使用账号 `zhangyi01` 密码 `ZYchandao2025!` 获取 Token

## MODIFIED Requirements

无

## REMOVED Requirements

无
