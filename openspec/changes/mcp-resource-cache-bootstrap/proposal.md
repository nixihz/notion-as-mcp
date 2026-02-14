## Why

Notion API 网络不稳定，导致 MCP server 的 resources 和 prompts 实时读取经常失败。需要引入缓存机制来提高可靠性，同时通过定时刷新保持数据时效性。

## What Changes

- 新增启动时预加载缓存：在 MCP server 启动时主动获取并缓存所有 resources 和 prompts
- 新增定时刷新机制：启动一个后台 goroutine，每 5 分钟自动异步获取最新数据
- 缓存更新策略：比较内容哈希，有变化才更新缓存
- 将现有的实时读取改为优先读取缓存，缓存未命中时回退到实时读取

## Capabilities

### New Capabilities
- `mcp-resource-cache`: MCP resources 的缓存管理（启动预热 + 定时刷新）
- `mcp-prompt-cache`: MCP prompts 的缓存管理（启动预热 + 定时刷新）

## Impact

- `internal/server/server.go`: 修改 server 初始化逻辑
- `internal/cache/`: 扩展现有缓存机制支持预热和定时刷新
- `cmd/serve.go`: 配置定时刷新参数
