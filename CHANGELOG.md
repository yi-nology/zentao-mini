# zentao-mini CHANGELOG

## [v1.9.0] — 2026-11-13

### 兼容清理：StdioTransport.Start() 退役

删除 `StdioTransport.Start()`（后台 goroutine 启动 stdio 监听的旧用法，
与 zentao-max v0.10.0 同口径）。库外部如有直接调用 `Start()` 的集成方，
改用 `Run(ctx)`（自行放 goroutine 即后台语义）。

配套接线调整（对外行为不变）：

- `Run(ctx)` 不再隐式调用 `SetTransport(TransportStdio)`，只负责监听
- 独立 stdio 入口（`cmd/mcp`）启动时自行 `SetTransport(TransportStdio)`
- both 模式的 HTTP 入口改为 `go transport.Run(ctx)` 启动后台 stdio 监听，
  传输模式仍标记 http（与 MCPModeManager「由启动入口单独指定」的设计一致）
