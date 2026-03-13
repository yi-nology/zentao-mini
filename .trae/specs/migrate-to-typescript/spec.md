# 前端 TypeScript 迁移规范

## Why
当前前端项目使用 JavaScript，缺乏类型安全保障，在开发过程中容易出现类型错误，降低了代码的可维护性和开发效率。迁移到 TypeScript 可以提供静态类型检查、更好的 IDE 支持和代码提示，提升代码质量和开发体验。

## What Changes
- 安装 TypeScript 相关依赖（typescript, @types/node, vue-tsc 等）
- 创建 TypeScript 配置文件（tsconfig.json, tsconfig.node.json）
- 将所有 .js 文件迁移为 .ts 文件
- 将 Vue 组件中的 `<script setup>` 改为 `<script setup lang="ts">`
- 为 API 响应数据添加类型定义
- 为 Vue Router 添加类型支持
- 更新 vite.config.js 为 vite.config.ts
- 为项目添加类型声明文件

## Impact
- Affected specs: 前端开发工作流
- Affected code: 
  - frontend/src/main.js → main.ts
  - frontend/src/router/index.js → index.ts
  - frontend/src/api/*.js → *.ts
  - frontend/vite.config.js → vite.config.ts
  - frontend/src/**/*.vue (所有 Vue 组件)

## ADDED Requirements

### Requirement: TypeScript 配置
系统应提供完整的 TypeScript 配置，支持 Vue 3 和 Vite 项目。

#### Scenario: TypeScript 编译成功
- **WHEN** 开发者运行 `npm run build` 或 `npm run dev`
- **THEN** TypeScript 编译器能够成功编译所有代码，无类型错误

### Requirement: 类型定义
系统应为所有 API 响应、组件 props 和 emits 提供完整的类型定义。

#### Scenario: API 响应类型安全
- **WHEN** 调用 API 获取数据
- **THEN** 返回的数据具有明确的类型定义，IDE 能够提供准确的代码提示

### Requirement: Vue 组件类型支持
所有 Vue 组件应使用 TypeScript 编写，提供 props 和 emits 的类型定义。

#### Scenario: 组件 props 类型检查
- **WHEN** 使用组件时传递 props
- **THEN** TypeScript 能够检查 props 类型是否正确

## MODIFIED Requirements

### Requirement: 开发脚本
更新 package.json 中的脚本命令，支持 TypeScript 编译检查。

**变更前**:
```json
{
  "scripts": {
    "dev": "vite",
    "build": "vite build"
  }
}
```

**变更后**:
```json
{
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc && vite build",
    "type-check": "vue-tsc --noEmit"
  }
}
```

## REMOVED Requirements
无移除的需求。
