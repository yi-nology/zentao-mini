# 工时统计页面 - 产品需求文档

## Overview
- **Summary**: 复刻 timelog.html 页面，实现工时统计分析功能，包括筛选、统计概览、图表展示和明细表格。
- **Purpose**: 提供工时数据的可视化分析，帮助用户了解项目工时分布和趋势。
- **Target Users**: 项目管理人员和开发团队成员。

## Goals
- 实现工时统计页面的完整功能
- 适应现有项目的 Vue 3 + Element Plus 风格
- 移除禅道认证部分
- 提供数据筛选、图表展示和明细表格功能

## Non-Goals (Out of Scope)
- 禅道认证功能（用户明确要求移除）
- 后端 API 开发（假设后端已提供）
- 数据持久化存储

## Background & Context
- 现有项目使用 Vue 3 + Element Plus 框架
- 项目结构包含 Layout 组件和多个视图组件
- 需要在现有菜单中添加工时统计选项
- 参考页面为 timelog.html，需要适配现有风格

## Functional Requirements
- **FR-1**: 提供时间范围快捷选择（本周、上周、本月、上月）
- **FR-2**: 提供产品、项目、执行/迭代、用户等筛选条件
- **FR-3**: 展示统计概览（总工时、工时记录数、涉及项目数、日均工时）
- **FR-4**: 展示每日工时图表
- **FR-5**: 展示按项目分布图表
- **FR-6**: 展示按任务类型分布图表
- **FR-7**: 展示工时流水明细表，支持排序和搜索

## Non-Functional Requirements
- **NFR-1**: 响应式设计，适配不同屏幕尺寸
- **NFR-2**: 页面加载和数据查询响应迅速
- **NFR-3**: 界面风格与现有项目保持一致
- **NFR-4**: 代码结构清晰，易于维护

## Constraints
- **Technical**: 使用 Vue 3 + Element Plus 框架
- **Dependencies**: 需要后端 API 支持工时数据查询

## Assumptions
- 后端已提供工时数据查询 API
- 现有项目的路由和菜单结构可以扩展

## Acceptance Criteria

### AC-1: 页面布局和风格
- **Given**: 用户访问工时统计页面
- **When**: 页面加载完成
- **Then**: 页面布局符合现有项目风格，包含侧边栏、头部和主内容区
- **Verification**: `human-judgment`

### AC-2: 筛选功能
- **Given**: 用户在工时统计页面
- **When**: 用户选择时间范围、产品、项目等筛选条件
- **Then**: 筛选条件正确应用，数据相应更新
- **Verification**: `programmatic`

### AC-3: 统计概览
- **Given**: 用户查询工时数据
- **When**: 数据加载完成
- **Then**: 统计概览卡片显示正确的总工时、记录数、项目数和日均工时
- **Verification**: `programmatic`

### AC-4: 图表展示
- **Given**: 用户查询工时数据
- **When**: 数据加载完成
- **Then**: 每日工时、项目分布、任务类型分布图表正确显示
- **Verification**: `programmatic`

### AC-5: 明细表格
- **Given**: 用户查询工时数据
- **When**: 数据加载完成
- **Then**: 工时流水明细表正确显示，支持排序和搜索
- **Verification**: `programmatic`

## Open Questions
- [ ] 后端 API 的具体接口和返回格式
- [ ] 数据源的具体字段和结构
- [ ] 是否需要添加导出功能