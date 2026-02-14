## ADDED Requirements

### Requirement: MCP prompts 缓存预热
MCP server 启动时 SHALL 预加载并缓存所有 prompt 页面，以提高 Notion API 不稳定时的可靠性。

#### Scenario: 服务器启动时缓存为空
- **WHEN** MCP server 启动
- **THEN** 系统从 Notion 获取所有 type="prompt" 的页面并存入缓存

#### Scenario: 缓存未命中时列出 prompts
- **WHEN** 列出 prompts 时缓存为空
- **THEN** 系统从 Notion 获取数据并填充缓存后再返回结果

### Requirement: Prompts 缓存定时刷新
MCP server SHALL 定期刷新缓存的 prompts 以保持数据相对新鲜。

#### Scenario: 定时器触发缓存刷新
- **WHEN** 5分钟定时器触发
- **THEN** 系统从 Notion 获取最新的 prompt 页面
- **AND** 使用内容哈希与缓存版本进行比较
- **AND** 如果不同，则用新内容更新缓存

#### Scenario: 刷新遇到速率限制
- **WHEN** 定时刷新触发 Notion API 速率限制
- **THEN** 系统记录警告并使用指数退避重试
- **AND** 继续使用现有缓存数据

### Requirement: 缓存优先的 prompt 读取
MCP server SHALL 在 prompt 操作时优先从缓存读取。

#### Scenario: 缓存包含有效数据
- **WHEN** prompt 数据存在于缓存中
- **THEN** 系统立即返回缓存数据，不调用 Notion API

#### Scenario: 缓存过期或不存在
- **WHEN** prompt 不在缓存中或缓存 TTL 已过期
- **THEN** 系统从 Notion 获取并更新缓存
