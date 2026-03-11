# 禅道SDK列表接口分页适配 - 产品需求文档

## Overview
- **Summary**: 为禅道SDK中所有列表接口添加分页查询支持，确保与更新后的禅道API保持一致。
- **Purpose**: 解决部分列表接口缺少分页支持的问题，提高API调用的灵活性和性能。
- **Target Users**: 开发者使用禅道SDK进行项目管理和开发。

## Goals
- 为所有列表接口添加分页查询参数（page和limit）
- 确保分页参数正确传递到禅道API
- 保持API接口的向后兼容性
- 提供清晰的文档和使用示例

## Non-Goals (Out of Scope)
- 不修改现有功能的实现逻辑
- 不添加新的功能特性
- 不改变返回数据结构

## Background & Context
- 禅道SDK已更新，所有列表接口现在都支持分页查询
- 目前builds.go中的GetBuildsByProject和GetBuildsByExecution函数未支持分页

## Functional Requirements
- **FR-1**: 为GetBuildsByProject函数添加分页参数（page和limit）
- **FR-2**: 为GetBuildsByExecution函数添加分页参数（page和limit）
- **FR-3**: 确保分页参数正确传递到API请求中
- **FR-4**: 保持函数返回类型不变，确保向后兼容

## Non-Functional Requirements
- **NFR-1**: 代码风格与现有代码保持一致
- **NFR-2**: 函数命名和参数顺序符合现有规范
- **NFR-3**: 错误处理逻辑与现有代码保持一致

## Constraints
- **Technical**: 基于Go语言实现，使用现有的HTTP客户端
- **Dependencies**: 依赖禅道API的分页参数支持

## Assumptions
- 禅道API已经支持所有列表接口的分页查询
- 分页参数的默认值可以设置为合理的默认值（如page=1, limit=50）

## Acceptance Criteria

### AC-1: GetBuildsByProject函数支持分页
- **Given**: 调用GetBuildsByProject函数时提供page和limit参数
- **When**: 执行API请求
- **Then**: API请求中包含正确的分页参数，返回对应页的数据
- **Verification**: `programmatic`

### AC-2: GetBuildsByExecution函数支持分页
- **Given**: 调用GetBuildsByExecution函数时提供page和limit参数
- **When**: 执行API请求
- **Then**: API请求中包含正确的分页参数，返回对应页的数据
- **Verification**: `programmatic`

### AC-3: 向后兼容性
- **Given**: 调用修改后的函数时不提供分页参数
- **When**: 执行API请求
- **Then**: 函数使用默认分页参数，正常返回数据
- **Verification**: `programmatic`

## Open Questions
- [ ] 分页参数的默认值应该设置为多少？
- [ ] 是否需要更新相关的类型定义？