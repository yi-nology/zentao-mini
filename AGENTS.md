# AGENTS.md — zentao-mini

## Project Overview

**Wails v3 alpha** desktop app (Go + Vue 3) for Zentao (禅道) project management. Also deployable as a standalone HTTP server or embedded app with static frontend.

**Go module name is `github.com/yi-nology/zentao-mini`**. Import paths use `github.com/yi-nology/zentao-mini/backend/...`.

> ⚠️ **Wails v3 still in alpha** (`v3.0.0-alpha2.117` as of 2026-07). API may churn between versions. Pin the version in `go.mod` rather than chasing `@latest`.

## Architecture

```
main.go + app.go          → Wails v3 entrypoint
                             - application.New + Service registration
                             - SystemTray (native, v3 新增)
                             - GlobalShortcut (CmdOrCtrl+Shift+Z 唤起)
                             - EventBus subscription → app.Event.Emit
backend/
  cmd/server/main.go      → HTTP server entrypoint (config-driven)
  cmd/app/main.go         → Embedded app entrypoint (static files baked in via ldflags)
  core/
    app/                  → Application interface, Wire DI
    event/                → In-process event bus (pub/sub)
    handlers/             → HTTP handlers (cache.go / logs.go)
    service/              → Business logic (cache_service / dashboard_service)
    storage/              → SQLite offline cache (modernc.org/sqlite, pure Go)
    logger/               → zap + ring_buffer
    metrics/              → Prometheus + cache hit rate
frontend/
  src/                    → Vue 3 + Element Plus + Chart.js
    composables/          → useTableColumns / useTheme / useDesktopNotification / useExternalLink
    utils/export.ts       → Excel/CSV/PDF exporter
  bindings/               → wails3 generate 输出（不要手改）
build/                    → v3 构建系统（Taskfile + 平台子目录）
Taskfile.yml              → v3 主构建入口
docs/grafana/             → Grafana dashboard JSON
```

Three runtime modes:
- **Wails desktop**: `wails3 task dev` / `make run`（启用 SystemTray / 全局快捷键 / 桌面通知）
- **HTTP server**: `cd backend && go run cmd/server/main.go`
- **Embedded app**: `cd backend && go run cmd/app/main.go`

## Commands

### Development
```bash
make run                  # wails3 task dev (hot reload frontend + Go)
make frontend-dev         # Vite dev server only
cd backend && make run    # HTTP server only
```

### Build
```bash
make build                # wails3 task build（当前平台）
make release              # wails3 task package（带打包）
wails3 task linux:build   # 指定平台（需在该平台环境或 Docker）
make wails-generate       # 重新生成前端 bindings（改 App 方法后必须）
```

### Test & Lint
```bash
cd backend && make check              # fmt + lint + test
cd backend && make test               # go test
cd frontend && npm run type-check     # vue-tsc --noEmit
cd frontend && npm run build          # vite build
```

### Cross-compilation
v3 + CGO 不再支持 macOS 上交叉编译 Linux/Windows。三个选项：
- **CI**: 推送到 master/tag，GitHub Actions 在对应平台 runner 构建
- **Docker**: `wails3 task setup:docker && wails3 task linux:build`
- **server-only**（无 GUI）: `cd backend && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./cmd/server`

## Key Conventions & Gotchas

- **Wire DI**: `backend/core/app/wire.go` 定义注入；改 providers 后 `wire ./backend/core/app/`
- **Frontend env switching**: 构建脚本 `cp .env.wails .env`，`.env` gitignored
- **Config priority**: AppConfig > database stored config > env vars (`ZENTAO_MINI_` prefix)
- **Default HTTP port**: `12345`
- **Generated code**: `frontend/bindings/` 是 wails3 generate 自动生成的，不要手改。改 `app.go` 的 App 方法后必须运行 `wails3 generate bindings -d frontend/bindings` 或 `make wails-generate`
- **Wails v3 Service 模式**: App 实现 `ServiceName/ServiceStartup/ServiceShutdown` 接口，通过 `application.NewService(app)` 包装注入到 `application.Options.Services`
- **Static files for embedded app**: `scripts/copy-static.sh` 复制 `frontend/dist/` → `backend/cmd/app/static/`
- **CI**: push to `master` 或 `v*.*.*` tag 触发，三平台 artifacts

## Feature Behaviors (mode-specific)

| Feature | Wails desktop | HTTP / Embedded |
|---------|--------------|-----------------|
| SystemTray（原生托盘）| ✅ v3 原生 | ❌ |
| GlobalShortcut（全局快捷键 CmdOrCtrl+Shift+Z）| ✅ v3 原生 | ❌ |
| Desktop notifications | ✅ app.Event.Emit | ❌（EventBus 无订阅者）|
| Hide-to-tray | ✅ | N/A |
| Offline SQLite cache | ✅ | ✅ |
| Log viewer / Grafana | ✅ | ✅ |

## Authentication Modes (双认证模式)

应用支持两种禅道认证与数据访问模式，按配置的 `realm` 字段自动切换：

| 模式 | 认证 | 数据访问 | 适用 |
|------|------|----------|------|
| **Token 模式**（默认） | 用户名/密码换 REST Token | `/api.php/v1/*` REST 端点（Token 头） | 标准禅道开源版/企业版 |
| **Session 模式** | `/user-login.html` PHP 会话 | `*.json` 端点（zentaosid cookie） | 禁用 REST API 的禅道实例（如麒麟 pm.kylin.com） |

**配置方式**：
- `~/.zentao-mini/auth.db` 里 `AuthConfig.Realm` 非空（如 `"kydc"`）→ Session 模式
- 前端登录表单选择认证域：`kydc`（麒麟 SSO，密码明文）/ `local`（禅道本地，md5 混淆）
- 环境变量：`ZENTAO_MINI_REALM` / `ZENTAO_REALM`

**实现关键文件**（`backend/core/zentao/`）：
- `client_auth_mode.go` — `AuthMode` / `RealmKylinSSO` 常量
- `client_session.go` — `SessionTransport`（登录、cookie jar、双层 JSON 解包、re-login 重试）
- `client_session_*.go` — 各实体的 `.json → SDK` 映射（bugs/stories/tasks/products/projects/executions/users/builds/efforts/cases/plans/programs/releases/testtasks/tickets/feedbacks）
- `client_writes*.go` — 写操作（bug/task/story/... 的 create/update/delete/resolve/close/start/finish 等），token + session 双模式
- `client.go` 的 `NewSessionClient` / `UpdateSessionConfig` / `doSessionJSON` / `doSessionPost` 做模式分发

**Gotchas**：
- Session 模式下 kylin SSO 短时间内频繁登录会被临时封账号（运营层限制），代码已加 30s 登录限流
- `.json` 端点每个实体的**分页 URL 段格式不同**（bug/project/execution 互不一致），见各 `client_session_*.go` 的注释
- 数值字段用 `json.Number` + `UseNumber()` 解码（兼容 int/string 混用）
- 未认证返回 HTTP 200 + `data.locate` 指向 `user-login`（不是 401），`IsSessionAuthError` 专门识别

## API Endpoints

**读**（GET）：`/api/{products,projects,executions,bugs,builds/{project,execution},stories,tasks,users,users/all,users/current,cases,plans,programs,releases,testtasks,tickets,feedbacks,timelog/{analysis,dashboard,efforts},dashboard,project/overview,personal/timelog,search}`

**写**（POST/PUT/DELETE）：
- Bug: `/api/bugs`（POST）、`/api/bugs/:id`（PUT/DELETE）、`/api/bugs/:id/{resolve,close,assign,confirm,activate}`（POST）
- Task: `/api/tasks`（POST）、`/api/tasks/:id`（PUT/DELETE）、`/api/tasks/:id/{start,finish,pause,assign,activate,effort}`（POST）
- Story: `/api/stories`（POST）、`/api/stories/:id`（PUT/DELETE）、`/api/stories/:id/change`（POST）
- Plan: `/api/plans`（POST）、`/api/plans/:id`（PUT/DELETE）、`/api/plans/:id/{link-stories,unlink-stories,link-bugs,unlink-bugs}`（POST）
- Case/Ticket/Feedback: `/api/{cases,tickets,feedbacks}`（POST）、`/:id`（PUT/DELETE），外加 `/api/feedbacks/:id/{assign,close}`（POST）
- Product/Project/Program/Execution/Build/User: `/api/{products,projects,programs,executions,builds,users}`（POST）、`/:id`（PUT/DELETE）

## Dependencies

- Go 1.25+（推荐 1.26），Node 24+
- Wails v3 CLI: `go install github.com/wailsapp/wails/v3/cmd/wails3@latest`
- `@wailsio/runtime` npm 包（已写入 package.json）
- golangci-lint
- Linux 完整构建需要: `libgtk-3-dev libwebkit2gtk-4.1-dev libayatana-appindicator3-dev pkg-config` 等
- macOS: Xcode Command Line Tools
- Windows: WebView2 Runtime（Win10/11 默认自带）

## Data files

- `~/.zentao-mini/cron.db` — JSON 存储定时任务和执行日志
- `~/.zentao-mini/cache.db` — SQLite 离线缓存（可安全删除）

## Known Issues / Migration Notes

- v3 alpha 阶段 API 偶有变化，升级版本前先 `wails3 doctor` 检查
- `WebviewWindowOptions.Hidden:true` 在 Windows 上有 bug ([#4498](https://github.com/wailsapp/wails/issues/4498))
- 全局快捷键在某些 Linux 桌面环境（如 Wayland）可能不工作
- 若 wails3 generate 失败，尝试 `wails3 generate bindings -clean -d frontend/bindings`


