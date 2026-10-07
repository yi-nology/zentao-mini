# zentao-mini REST API 参考

基底 URL：`http://<host>:12345`，业务端点在 `/api` 前缀下。
成功响应统一为 `{"code":200,"message":"success","data":...}`；列表为 `{"code":200,"data":{"list":[...],"total":N,"page":P,"pageSize":S}}`。
错误码：`40000` 参数错误、`40100` 未登录（写操作）、`50002` 禅道服务调用失败。

## 认证

- **读（GET）**：匿名可用。
- **写（POST/PUT/DELETE）**：需管理员。`POST /api/auth/login` `{"username","password"}` → `data.token`，后续请求带 `Authorization: Bearer <token>`（网页端自动用 HttpOnly Cookie）。
- 限流：默认 600 次/分钟/IP，超限 HTTP 429；`X-RateLimit-Limit/Remaining/Reset` 响应头可读。
- MCP 端点 `/mcp*` 独立鉴权（见 references/mcp.md），不走平台账号。

## 查询端点（GET，匿名）

| 端点 | 关键参数 | 说明 |
|------|---------|------|
| `/api/products` | 无（自动翻页全量） | 产品列表，`status`: normal/closed |
| `/api/projects` | `productId` | 项目列表 |
| `/api/executions` | `productId`, `projectId` | 执行/迭代列表 |
| `/api/bugs` | `productId`, `status`, `assignedTo`, `openedBegin`/`openedEnd`, `type`, `page`, `pageSize` | Bug 列表 |
| `/api/stories` | `productId`, `projectId`, `executionId`, `status`, `assignedTo`, `page`, `pageSize` | 需求列表 |
| `/api/tasks` | `productId`, `projectId`, `executionId`, `status`, `assignedTo`, `page`, `pageSize` | 任务列表 |
| `/api/users/all` | 无 | 全量用户 |
| `/api/users/current` | 无 | 当前禅道账号 |
| `/api/dashboard` | `productId`（可选） | 仪表盘聚合：bug/需求/任务计数与多维分布 |
| `/api/timelog/analysis` | `productId`, `dateFrom`, `dateTo`, `assignedTo` | 工时分析（大产品 20-45s，有缓存） |
| `/api/timelog/dashboard` | `productId`, `dateFrom`, `dateTo` | 工时看板汇总 |
| `/api/timelog/efforts` | `taskId` | 单任务工时明细 |
| `/api/personal/timelog` | `dateFrom`, `dateTo` | 个人工时 |
| `/api/project/overview` | `projectId` | 项目概览 |
| `/api/search` | `keyword` | 全局搜索 Bug/需求/任务 |
| `/api/healthz` | 无 | 心跳：禅道连通性 + 各数据面自检 |
| `/api/logs` | `page`, `pageSize`, `level`, `keyword` | 系统日志（环形缓冲） |
| `/api/logs/status` | 无 | 日志缓冲水位 |
| `/api/cache/status` | 无 | 离线缓存状态 |

`page` 从 1 起，`pageSize` 默认 20、上限 100。

## 写端点（POST/PUT/DELETE，需管理员 Bearer）

| 端点 | 说明 |
|------|------|
| `POST /api/auth/login` / `POST /api/auth/logout` | 登录/登出（本身匿名可用） |
| `GET /api/auth/status` | 当前平台角色（anonymous=readonly） |
| `POST /api/scheduler/tasks` | 新建定时报告任务 |
| `PUT /api/scheduler/tasks/:id` | 修改任务 |
| `DELETE /api/scheduler/tasks/:id` | 删除任务 |
| `POST /api/scheduler/tasks/:id/run` | 立即执行 |
| `POST /api/scheduler/preview` | 报告预览（匿名可用，只读语义）。`reportType`：`bug`/`requirement`/`task`/`bug-aging`（配 `agingDays`）/`daily-report-check`（配 `checkHours` 每工作日工时阈值，默认 8；`period`=`YYYY-MM` 可选，为周期截止月，缺省检查上月16日～本月15日） |
| `POST /api/scheduler/test-webhook` | 测试 webhook 推送 |
| `GET /api/scheduler/tasks/:id/logs`、`GET /api/scheduler/logs` | 任务执行日志 |
| `DELETE /api/logs` | 清空日志缓冲 |
| `DELETE /api/cache`、`DELETE /api/cache/:entityType` | 清空/失效离线缓存 |
| `POST /api/v1/init/upload` | 上传加密禅道配置（仅首次启动可匿名） |

### Bug 写操作（需管理员 Bearer）

| 端点 | 说明 |
|------|------|
| `POST /api/bugs/:id/comments` | 给 Bug 添加备注/评论，body `{"comment":"..."}`（必填） |
| `POST /api/bugs/:id/transitions` | Bug 状态流转，body 见下 |

`transitions` 请求体：

```json
{
  "action": "confirm | resolve | close | activate | assign",
  "resolution": "resolve 必填：fixed/bydesign/duplicate/notrepro/postponed/willnotfix/external",
  "resolvedBuild": "resolve 建议填，如 trunk 或版本号",
  "assignedTo": "assign 必填；activate 可选（重新指派），值为禅道账号",
  "comment": "任意动作可附带备注，记入操作历史（可选）"
}
```

- 流转与状态机的匹配由禅道侧校验（如 `resolved` 状态不能再 `resolve`），禅道的报错会原样返回（code 50002）。
- 兼容性：不同禅道版本的 REST v1 能力不一（实测 pm.kylin.com 无 `activate` 端点、`close` 偶发静默失效）。`activate`/`close` 在 REST 失败或未生效时自动降级到 Web 会话通道（与评论同机制），新老版本禅道均可用。
- 两个接口成功后 `data` 都返回流转后的 Bug 最新快照。
- 评论以禅道配置账号名义发表（Web 会话通道），Bug 历史中出现 `Commented` 记录，与网页端手动评论一致。

## 字段约定

- 所有 ID 参数均为**字符串**形式的数字（`"1029"`）。
- Bug 的 `severity` 1-5（1 致命），`pri` 1-4；`assignedTo` 为对象 `{id, account, realname}`。
- 任务 `status`: wait/doing/done/closed/pause/cancel；需求 `status`: active/developing/testing/closed。
- 产品 `status`: normal（正常）/closed（关闭）。

## curl 示例

```bash
# 登录拿 token
TOKEN=$(curl -s -X POST http://localhost:12345/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | python3 -c "import json,sys;print(json.load(sys.stdin)['data']['token'])")

# 管理员写操作示例：清空日志
curl -s -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:12345/api/logs

# Bug 写操作：加评论
curl -s -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"comment":"已定位，今天发版修复"}' http://localhost:12345/api/bugs/456/comments

# Bug 写操作：解决 Bug
curl -s -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"action":"resolve","resolution":"fixed","resolvedBuild":"trunk","comment":"见提交 abc123"}' \
  http://localhost:12345/api/bugs/456/transitions

# 分页查活跃 Bug（第 2 页）
curl -s "http://localhost:12345/api/bugs?productId=1029&status=active&page=2&pageSize=20"
```
