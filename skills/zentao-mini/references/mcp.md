# zentao-mini MCP 参考

## 运行形态

| 形态 | 入口 | 适用 |
|------|------|------|
| HTTP（随主服务） | `http://<host>:12345/mcp` | AI Agent、脚本、远程调用 |
| stdio（独立进程） | `zentao-mini-mcp` 二进制（编译自 `backend/cmd/mcp`） | Claude Desktop、Cursor、Claude Code 等标准 MCP 客户端 |

两种形态暴露相同的 9 个工具；HTTP 形态受配置 `mcp.enabled`（默认 true）控制。

## 协议（仅 MCP 2026-07-28，不支持旧版本）

服务端**只支持最新协议 2026-07-28**，无握手、无状态：

- 每个请求必须在 `params._meta` 携带 `"io.modelcontextprotocol/protocolVersion": "2026-07-28"`，缺失或不支持返回错误码 **-32022**（`error.data` 带正确版本指引）
- `server/discover`：唯一免版本检查的方法，返回支持的版本、能力与服务身份；客户端可先调它做连通性探测
- `tools/list`：结果带 `resultType:"complete"` 与缓存提示（`ttlMs=300000, cacheScope:"private"`）；工具带 `title` 与 `annotations`（查询工具只读，Bug 写工具 `readOnlyHint:false`）
- `tools/call`：结果带 `resultType:"complete"`；对象结果附带 `structuredContent`（文本 content 同时保留）
- 已移除的方法：`initialize`、`notifications/initialized`、`ping`（旧客户端调用会收到 -32022 / -32601）
- JSON-RPC 通知（无 id）一律不响应

```bash
# 标准调用示例（stdio 单行 / HTTP POST body 同构）
{"jsonrpc":"2.0","id":1,"method":"tools/call",
 "params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28"},
           "name":"get_bugs","arguments":{"productId":"1029"}}}
```

两套消息协议自动识别：

1. **标准 MCP JSON-RPC 2.0**（如上，仅 2026-07-28）
2. **zentao-mini 简化 action 协议**（自有便捷协议，非 MCP 版本，兼容保留）：`{"action":"get_bugs","params":{...}}` 或 `GET /mcp/<action>?param=value`

HTTP 传输注意：
- JSON-RPC 批量请求（数组 body）不支持，返回 `-32600`
- `GET /mcp`（无 action、SSE Accept）返回 405（本服务无服务端推送）；带 `action=` 的旧便捷查询不受影响

## 访问控制（分级：无 token 只读，有 token 读写）

配置段 `mcp`：

```yaml
mcp:
  enabled: true          # 总开关
  transport: http        # http | stdio | both（主服务内）
  read_only: false       # true 时即使持有效 token 也拒绝写（服务级只读策略，403）
  token: ""              # 写操作凭证；env: ZENTAO_MINI_MCP_TOKEN 或 MCP_TOKEN
  # actions: [...]       # action 白名单（可选，读写均生效）
```

**分级权限模型**：

| 操作 | 无 token / 未配置 token | 有效 token |
|------|------------------------|-----------|
| 读（get_products / get_bugs / get_stories / get_tasks / get_users / get_timelog） | ✅ 匿名可用 | ✅ |
| 写（add_bug_comment / transition_bug） | ❌ 403（匿名只读） | ✅ |

- HTTP Token 携带：`Authorization: Bearer <token>`（优先）或 `?token=<token>`；**仅写操作需要**
- token 错误：401；`read_only=true`：写一律 403（即使 token 正确）
- stdio 子进程：Token 经客户端配置的 env（`ZENTAO_MINI_MCP_TOKEN`）注入，进程持有即可写
- MCP 端点独立于平台账号体系（平台管理员账号不适用于 MCP 调用）

## HTTP 端点

| 方法+路径 | 说明 |
|-----------|------|
| `GET /mcp/tools` | 工具发现：`{"status":"ok","count":11,"tools":[{name,description,inputSchema}...]}` |
| `GET /mcp/tools/:name` | 单个工具的 JSON Schema |
| `GET /mcp/<action>?param=value` | 按路径调用（action 别名见下） |
| `GET /mcp?action=<action>&param=value` | 统一查询入口 |
| `POST /mcp` | `{"action":"...","params":{...}}`（标准方式） |
| `GET/POST /mcp/ping` 等 | 每个 action 同时注册了 `/mcp/<别名>`（products/bugs/...） |

action 别名：`ping` `products` `projects` `executions` `bugs` `stories` `tasks` `users` `timelog`。

参数也可用 query 传递（GET 自动收集）：`productId` `projectId` `executionId` `status` `assignedTo` `dateFrom` `dateTo` `page` `pageSize`。

## 响应格式

- `tools/call` 返回 `{"content":[{"type":"text","text":"<工具结果 JSON>"}]}`，`text` 反序列化后多数为 `{"status":"ok","data":...}`（判断 `status` 后取 `data`；Bug/需求/任务列表在 `data.list`，分页字段 `data.total`）
- 简化 action 协议：多数返回 `{"status":"ok","data":...}`；`get_products` 等直接返回数组
- 错误：JSON-RPC `tools/call` 的工具级错误在 `result.isError=true` + `content[0].text`；协议级错误为标准 `error` 对象（-32601 方法不存在等）
- 服务关闭：HTTP 503 `{"status":"error","message":"MCP service is disabled"}`
- HTTP Token 错误：401；action 不在白名单/只读拦截：403

## 工具 Schema 速查

### get_products
无参数。返回产品数组：`[{id, name, code, type, status(normal|closed), desc}]`

### get_projects
- `productId` (string, 可选)

### get_executions
- `projectId` (string, 可选)
- `productId` (string, 可选)

### get_bugs
- `productId` (string, 可选)
- `status` (string, 可选): `active` | `resolved` | `closed`
返回 `{"status":"ok","data":{"list":[Bug], "total":N}}`；Bug：`{id, title, severity(1-5), pri, status, type, assignedTo:{account,realname}, openedDate, steps}`

### get_stories
- `productId` (string, 可选)
返回需求：`{id, title, status(active|developing|testing|closed), stage, pri, estimate, assignedTo, openedDate}`

### get_tasks
- `productId` (string, 可选)
- `executionId` (string, 可选)
返回任务：`{id, name, status(wait|doing|done|closed|pause|cancel), pri, estimate, consumed, left, assignedTo, deadline}`

### get_users
无参数。`[{id, account, realname, role, email, status}]`

### get_timelog
- `productId` (string, 可选)
- `dateFrom` / `dateTo` (string, YYYY-MM-DD, 可选)
返回：`{taskCount, effortCount, totalHours, byDate:[{date,hours,count}], byProject:[{name,hours,count}], byType:[{name,hours,count}], efforts:[...]}`

> 大产品的工时统计涉及逐任务扇出查询，真实环境可能耗时 20-45 秒；服务端有约 2 分钟缓存，重复调用很快。

### ping
无参数。`{"message":"Pong","status":"ok","version":"1.0"}`

### add_bug_comment（写）
- `bugId` (string, 必填)
- `comment` (string, 必填，支持禅道富文本 HTML)
给 Bug 添加备注。以禅道配置账号名义发表，Bug 历史出现 `Commented` 记录，与网页端评论一致。成功返回 `{"status":"ok","data":<Bug 最新快照>}`。

### transition_bug（写）
- `bugId` (string, 必填)
- `action` (string, 必填): `confirm`(确认) | `resolve`(解决) | `close`(关闭) | `activate`(激活/重开) | `assign`(指派)
- `resolution` (string, resolve 必填): `fixed` | `bydesign` | `duplicate` | `notrepro` | `postponed` | `willnotfix` | `external`
- `resolvedBuild` (string, resolve 建议填，如 trunk)
- `assignedTo` (string, assign 必填；activate 可选，重新指派，值为禅道账号)
- `comment` (string, 可选，随流转记入操作历史)
成功返回 `{"status":"ok","message":"状态流转成功(...)","data":<Bug 最新快照>}`。
状态机合法性由禅道校验（如 resolved 不能再 resolve），业务错误原样透出。

> 写工具与查询工具共用 mcp.token 鉴权；`mcp.read_only: true` 时返回 403 拒绝。
> 等价的 REST 接口见 references/api.md 的 "Bug 写操作" 段（需平台管理员 Bearer）。
> 兼容性：老版本禅道 REST v1 可能缺 `activate` 端点或 `close` 静默失效，服务端会自动降级到 Web 会话通道完成操作。

## 桌面客户端配置

前置：编译 stdio 二进制 `cd backend && go build -o zentao-mini-mcp ./cmd/mcp`；子进程需能读到禅道连接配置（`~/.zentao-mini/auth.db` + 环境变量 `ZENTAO_MINI_SECURITY_ENCRYPTION_KEY`）或用 `ZENTAO_SERVER/ZENTAO_ACCOUNT/ZENTAO_PASSWORD` 直接指定。

### Claude Desktop / Cursor（`mcpServers`）

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

### Claude Code

```bash
claude mcp add zentao-mini -- /path/to/zentao-mini-mcp
claude mcp list   # 验证
```

### 远程 HTTP（Agent/脚本直连已部署服务）

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
