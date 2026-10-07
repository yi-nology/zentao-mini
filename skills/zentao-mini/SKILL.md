---
name: zentao-mini
description: 对接 zentao-mini（禅道项目管理数据的轻量网关）查询禅道产品/Bug/需求/任务/工时数据。当用户需要查询禅道项目管理数据、统计 Bug 或工时、或要把 AI 客户端接入禅道时使用。支持 MCP 工具协议与 REST API 两种方式。
---

# zentao-mini 对接指南

zentao-mini 是禅道（ZenTao）项目管理数据的轻量网关：连接禅道服务器后，对外提供 MCP 工具与 REST API。**查询匿名可用**；Bug 写操作（评论/状态流转）走 MCP 写工具（受 mcp.token 与 read_only 配置约束）或 REST 写接口（需平台管理员账号）。

## 第一步：确认服务与连通性

向用户确认服务地址（本机开发默认 `http://localhost:12345`），然后验证连通：

```bash
curl -s http://localhost:12345/mcp/tools
# → {"status":"ok","count":11,"tools":[...]}   即服务可用
```

失败时检查：服务是否启动（`curl http://localhost:12345/health`）、MCP 是否被关闭（配置 `mcp.enabled`）。

## 第二步：选择对接方式

| 方式 | 适用场景 | 入口 |
|------|---------|------|
| **MCP HTTP**（推荐给 AI Agent） | 你能发 HTTP 请求，想用统一的工具协议 | `GET/POST /mcp` |
| **MCP stdio**（推荐给桌面客户端） | Claude Desktop / Cursor 等，配置一次长期使用 | `zentao-mini mcp` 子命令 |
| **REST API** | 需要分页、精细筛选，或宿主环境无 MCP | `GET /api/*` |

---

## 方式 A：MCP HTTP 调用

### 可用工具（11 个：9 只读查询 + 2 Bug 写操作）

| 工具 | 说明 | 参数（均可选） |
|------|------|--------------|
| `ping` | 连通性测试 | - |
| `get_products` | 产品列表 | - |
| `get_projects` | 项目列表 | `productId` |
| `get_executions` | 执行/迭代列表 | `projectId`, `productId` |
| `get_bugs` | Bug 列表 | `productId`, `status`(active/resolved/closed) |
| `get_stories` | 需求列表 | `productId` |
| `get_tasks` | 任务列表 | `productId`, `executionId` |
| `get_users` | 用户列表 | - |
| `get_timelog` | 工时统计 | `productId`, `dateFrom`, `dateTo`(YYYY-MM-DD) |
| `add_bug_comment` **（写）** | 给 Bug 加备注/评论 | `bugId`, `comment`（均必填） |
| `transition_bug` **（写）** | Bug 状态流转 | `bugId`, `action`(confirm/resolve/close/activate/assign) 必填；`resolution`(resolve 必填), `resolvedBuild`, `assignedTo`(assign 必填), `comment` 可选 |

写工具调用格式同样用 `POST /mcp`（`{"action":"add_bug_comment","params":{"bugId":"456","comment":"已定位"}}`）；服务端 `mcp.read_only: true` 或 token 不符时被拒（403/401）。REST 等价接口见 references/api.md。

### 调用格式（三选一）

```bash
# 1. GET 路径别名（最简洁）
curl -s "http://localhost:12345/mcp/bugs?productId=1029&status=active"

# 2. GET 统一入口
curl -s "http://localhost:12345/mcp?action=get_bugs&productId=1029"

# 3. POST JSON（参数复杂时）
curl -s -X POST http://localhost:12345/mcp \
  -H "Content-Type: application/json" \
  -d '{"action":"get_timelog","params":{"productId":"1029","dateFrom":"2026-09-01","dateTo":"2026-10-07"}}'
```

### 响应格式

- 多数工具返回 `{"status":"ok","data":...}`，**判断 `status=="ok"` 后取 `data`**；
- 部分工具（如 `get_products`）直接返回裸数组；
- 出错时 `{"status":"error","message":"..."}`（HTTP 400）。

### 鉴权（分级：无 token 只读，有 token 读写）

- **读工具**：永远匿名可用，不需要 token
- **写工具**（add_bug_comment / transition_bug）：需要服务端配置 `mcp.token` 且调用时携带：

```bash
-H "Authorization: Bearer <ZENTAO_MINI_MCP_TOKEN>"
# 或 ?token=<ZENTAO_MINI_MCP_TOKEN>
```

无凭证调用写工具：服务端未配置 token 时返回 403（提示配置以启用写）；已配置但凭证不符返回 401。服务端 `mcp.read_only: true` 时写一律 403。

### 标准 MCP 客户端（仅支持协议 2026-07-28）

服务端**只支持最新 MCP 协议 2026-07-28**（无握手、每请求 `_meta` 携带版本，旧版本会被 -32022 拒绝）。两种接入：

**stdio（本地子进程，需 zentao-mini-mcp 独立二进制，编译自 `backend/cmd/mcp`）**：

```json
{
  "mcpServers": {
    "zentao-mini": {
      "command": "/path/to/zentao-mini-mcp",
      "args": [],
      "env": { "ZENTAO_MINI_SECURITY_ENCRYPTION_KEY": "<部署加密密钥>" }
    }
  }
}
```

**HTTP（远程服务直连）**：

```json
{
  "mcpServers": {
    "zentao-mini": {
      "type": "http",
      "url": "http://<host>:12345/mcp",
      "headers": { "Authorization": "Bearer <ZENTAO_MINI_MCP_TOKEN>" }
    }
  }
}
```

注意：stdio 子进程需要能读到禅道连接配置（`~/.zentao-mini/auth.db` + 对应加密密钥），否则启动后工具返回空数据。客户端需支持 2026-07-28 无状态语义（每请求 `_meta` 携带版本）；仅支持旧握手（initialize）的客户端无法连接，可先用 `server/discover` 探测。

---

## 方式 B：REST API

基底路径 `/api`，查询类 GET 匿名可用。**列表接口统一返回** `{"code":200,"data":{"list":[...],"total":N,"page":P,"pageSize":S}}`。

```bash
# 分页查 Bug（page 从 1 起，pageSize 上限 100）
curl -s "http://localhost:12345/api/bugs?productId=1029&status=active&page=1&pageSize=20"

# 仪表盘聚合（bug/需求/任务计数与分布，适合概览提问）
curl -s "http://localhost:12345/api/dashboard?productId=1029"

# 工时分析（大产品可能耗时 20-45s，有 2 分钟缓存）
curl -s "http://localhost:12345/api/timelog/analysis?productId=1029&dateFrom=2026-09-01&dateTo=2026-10-07"
```

常用端点速查：`/api/products` `/api/projects` `/api/executions` `/api/bugs` `/api/stories` `/api/tasks` `/api/users/all` `/api/timelog/analysis` `/api/dashboard` `/api/search?keyword=`。完整参数见 `references/api.md`。

**写操作**（定时任务、清空日志等）需管理员：`POST /api/auth/login` 得 Bearer token 后携带。匿名调用写接口返回 401 `{"code":40100,...}`。

---

## 典型工作流

### 查"某产品当前活跃 Bug"
1. `get_products` 找到产品名 → 拿到 `productId`
2. `get_bugs?productId=<id>&status=active`
3. 按 `severity`/`assignedTo.account` 归纳汇报

### 查"某迭代任务进度"
1. `get_projects?productId=<id>` → `get_executions?projectId=<pid>` 拿迭代
2. `get_tasks?executionId=<eid>` → 按 `status`（wait/doing/done/closed）统计

### 查"团队最近工时投入"
`get_timelog?productId=<id>&dateFrom=&dateTo=` → `data.byDate`（按日）、`byProject`（按项目）、`byType`（按类型）

## 错误与限制

| 现象 | 含义 | 处理 |
|------|------|------|
| HTTP 400 `{"status":"error"}` | 参数错误/禅道不可达 | 看 `message` |
| HTTP 401 `code:40100` | 写操作未登录 | 走管理员登录 |
| HTTP 429 | 触发限流（默认 600/min） | 读 `X-RateLimit-Reset` 退避 |
| `data` 全为 0/空 | 产品 ID 不对或无数据 | 先 `get_products` 核对 ID |

**约定**：所有 ID 参数传**字符串**（`"1029"`）；`productId` 是最常用的过滤入口，不确定时先查产品列表。

## 参考文件

- `references/api.md` — REST API 全量端点、参数、响应结构
- `references/mcp.md` — MCP 工具 JSON Schema、客户端（Claude/Cursor 等）配置片段
