# AGENTS.md — zentao-mini

## Project Overview

**纯网页版** Web app (Go + Vue 3) for Zentao (禅道) project management. 两种运行形态:standalone HTTP server(`backend/cmd/server`)或内嵌前端的 embedded app(`backend/cmd/app`,go:embed)。桌面版(Wails)已于 v1.4.0 移除。

**Go module name is `github.com/yi-nology/zentao-mini`**. Import paths use `github.com/yi-nology/zentao-mini/backend/...`.


## Architecture

```
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
build/                    → v3 构建系统（Taskfile + 平台子目录）
Taskfile.yml              → v3 主构建入口
docs/grafana/             → Grafana dashboard JSON
```

Three runtime modes:
- **HTTP server**: `cd backend && go run cmd/server/main.go`
- **Embedded app**: `cd backend && go run cmd/app/main.go`

## Commands

### Development
```bash
cd frontend && npm run dev   # Vite dev server (端口 6100, /api 代理到 12345)
cd backend && go run cmd/server/main.go   # HTTP server (端口 12345)
```

### Build
```bash
cd frontend && npm run build && ./scripts/copy-static.sh
cd backend && CGO_ENABLED=0 go build -o zentao-mini ./cmd/app   # 内嵌前端单二进制
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
- **server-only**（无 GUI）: `cd backend && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./cmd/server`

## Key Conventions & Gotchas

- **Wire DI**: `backend/core/app/wire.go` 定义注入；改 providers 后 `wire ./backend/core/app/`
- **Frontend env switching**: `.env` 由构建脚本按需切换,gitignored;`.env.docker` 供镜像构建使用
- **Config priority**: AppConfig > database stored config > env vars (`ZENTAO_MINI_` prefix)
- **Default HTTP port**: `12345`
- **Static files for embedded app**: `scripts/copy-static.sh` 复制 `frontend/dist/` → `backend/cmd/app/static/`（CI 内自动执行）
- **CI**: push to `master` 或 `v*.*.*` tag 触发，三平台 artifacts

## Feature Behaviors

- 全部形态均为网页版;SystemTray/全局快捷键/桌面事件推送(Wails)已移除。
- 通知:应用内 ElNotification + 浏览器 Notification API(设置页开关)。

## Dependencies

- Go 1.25+（推荐 1.26），Node 24+
- golangci-lint
- Linux 完整构建需要: `libgtk-3-dev libwebkit2gtk-4.1-dev libayatana-appindicator3-dev pkg-config` 等
- macOS: Xcode Command Line Tools
- Windows: WebView2 Runtime（Win10/11 默认自带）

## Data files

- `~/.zentao-mini/cron.db` — JSON 存储定时任务和执行日志
- `~/.zentao-mini/cache.db` — SQLite 离线缓存（可安全删除）
- `~/.zentao-mini/auth.db` — 加密的禅道连接配置（默认主目录，可通过 `auth.db_path` 覆盖）

## Platform auth & MCP skills

- **平台访问控制**（`backend/core/auth`）：匿名只读（GET 放行），登录管理员读写。凭据 `auth.admin.username/password`（默认 admin/admin，env `ZENTAO_MINI_AUTH_ADMIN_USERNAME/PASSWORD`）。MCP 端点 `/mcp*` 豁免平台认证（自治 mcp.token）。
- **Bug 写能力**（2026-10-07）：`POST /api/bugs/:id/comments`（纯评论：禅道 v1 API 无备注端点，走 Web 会话通道复刻 `action-comment` 表单）与 `POST /api/bugs/:id/transitions`（confirm/resolve/close/activate/assign，v1 API，均可附 comment）；均需平台管理员登录，成功返回 Bug 最新快照。MCP 写工具 `add_bug_comment` / `transition_bug`（annotations readOnlyHint=false；`mcp.read_only` 与 `IsWriteAction` 拦截）。新增写动作时同步维护 tools.go 的 writeTools 与 server.go 的 IsWriteAction。
- **标准 MCP 协议**（`backend/core/mcp/jsonrpc.go`）：仅支持**最新协议 2026-07-28**（无握手、每请求 `params._meta` 携带版本，缺失/旧版本返回 -32022；`server/discover` 通告；stdio 与 `POST /mcp` 双通道）。不提供旧协议版本兼容。zentao-mini 自有 `{"action":...}` 简化协议兼容保留。stdio 日志强制走 stderr 不污染协议流。
- **对接 Skill**：`skills/zentao-mini/`（SKILL.md + references/api.md + references/mcp.md）——可直接分发或放入 Agent 的技能目录，教 AI 客户端对接禅道数据。改 MCP 工具/REST 后同步更新。

## Known Issues / Migration Notes



