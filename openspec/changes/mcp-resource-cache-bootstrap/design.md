## Context

当前 MCP server 对 Notion resources 和 prompts 采用实时读取策略，每次请求都直接调用 Notion API。Notion API 网络不稳定，导致频繁失败。用户希望引入缓存机制提高可靠性，同时保持数据时效性。

## Goals / Non-Goals

**Goals:**
- 提高 MCP server 可靠性，减少因网络问题导致的失败
- 启动时预加载缓存，减少首次请求延迟
- 定时刷新缓存，保持数据相对新鲜（5分钟间隔）
- 智能更新：仅当内容变化时才更新缓存

**Non-Goals:**
- 不改变 MCP 协议的接口语义
- 不做全量数据同步（只同步 resources 和 prompts）
- 不实现分布式缓存（单实例场景）

## Decisions

### Decision 1: 缓存存储方式

**选用**: 复用现有的两层缓存机制（内存 + 文件）

**理由**: 项目已有 `internal/cache/` 模块，内存缓存用于快速访问，文件缓存用于持久化。只需扩展其支持预热和定时刷新。

### Decision 2: 定时刷新实现

**选用**: Go `time.Timer` + goroutine

**理由**:
- 轻量级，无需额外依赖
- 可独立控制刷新间隔
- 优雅关闭（通过 context 取消）

**其他方案 considered**:
- `cron` 包: 过于重量，且我们需要相对简单的间隔执行
- 外部消息队列: 过度设计

### Decision 3: 变更检测策略

**选用**: 内容哈希比较（SHA256）

**理由**:
- Notion API 每次返回可能包含 `last_edited_time`，可以此判断
- 如果没有版本信息，则对整个响应 body 做哈希

### Decision 4: 缓存读取策略

**选用**: 缓存优先，回退实时读取

**理由**:
- 保证始终有数据可用（缓存或实时）
- 避免破坏现有错误处理逻辑
- 首次启动时缓存为空，直接实时读取并填充缓存

## Risks / Trade-offs

1. **缓存数据过期** → 通过 5 分钟定时刷新缓解
2. **启动时预热延迟** → 异步预热，不阻塞 server 启动完成信号
3. **Notion API 限流** → 定时任务需捕获限流错误，避免重试风暴

## Migration Plan

1. 扩展 `internal/cache/` 模块，添加 `Warm()` 和 `StartPeriodicRefresh()` 方法
2. 修改 `internal/server/server.go`，在 `Run()` 启动时调用缓存预热
3. 添加配置项 `CACHE_REFRESH_INTERVAL`（默认 5m）
4. 部署后观察日志确认缓存正常工作

## Open Questions

- 是否需要暴露缓存状态到 MCP server 的 tools？（用于调试）
- 定时刷新失败时是否需要告警？
