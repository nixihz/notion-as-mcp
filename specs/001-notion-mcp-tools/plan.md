# Implementation Plan: Notion MCP Server

**Branch**: `001-notion-mcp-tools` | **Date**: 2025-12-25 | **Spec**: [Link](spec.md)
**Input**: Feature specification from `/specs/001-notion-mcp-tools/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

构建一个 MCP 服务器，将 Notion 数据库作为数据源，通过类型字段区分 prompt/resource/tool 三种 MCP 原语。工具从 Notion 代码块中提取可执行代码。Go 1.25 开发，使用官方 go-sdk，Cobra CLI，env 配置，stdio 传输，带缓存和 slog 日志。

## Technical Context

**Language/Version**: Go 1.25
**Primary Dependencies**:
- `github.com/modelcontextprotocol/go-sdk` - MCP 协议实现
- `github.com/spf13/cobra` - CLI 框架
- `github.com/joho/godotenv` - 环境变量加载
- `log/slog` - Go 1.25+ 标准库结构化日志
**Storage**: 文件系统缓存（$HOME/.cache/notion-mcp/）+ 内存缓存
**Testing**: `testing` 标准库 + testify
**Target Platform**: macOS/Linux (本地部署)
**Project Type**: 单项目 CLI 工具
**Performance Goals**:
- 缓存响应 < 100ms
- 未缓存响应 < 2s
- 工具调用成功率 > 95%
**Constraints**:
- stdio 传输，无 Web 界面
- 本地部署
- 单用户场景
**Scale/Scope**:
- 单数据库
- 单用户
- 数百条目

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. 代码质量 ✓ PASS

| Requirement | Implementation |
|-------------|----------------|
| 单一职责 | 分离 CLI、服务、缓存、Notion 客户端职责 |
| 类型安全 | Go 静态类型，定义明确的数据结构 |
| 错误处理 | 所有错误路径处理，返回用户友好的消息 |
| 无调试代码 | 使用 slog 日志，移除所有 fmt.Print |

### II. 测试标准 ✓ PASS

| Requirement | Implementation |
|-------------|----------------|
| 单元测试 | 核心逻辑（缓存、解析、过滤）单元测试 |
| 契约测试 | MCP 协议消息格式验证 |
| 集成测试 | 完整用户旅程测试（mock Notion API） |
| 覆盖率 | 核心业务逻辑 > 80% |

### III. 用户体验一致性 ✓ PASS

| Requirement | Implementation |
|-------------|----------------|
| 统一交互模型 | 所有工具通过统一接口暴露 |
| 可预测反馈 | 结构化日志，清晰的错误信息 |
| 配置一致性 | 环境变量命名规范 |

### IV. 可维护性 ✓ PASS

| Requirement | Implementation |
|-------------|----------------|
| 模块化 | 清晰的包结构：cmd, internal/{server,cache,notion,tools} |
| 配置优先 | 环境变量驱动所有行为变化 |
| 依赖管理 | go.mod 锁定版本，定期安全扫描 |

### V. 文档 ✓ PASS

| Requirement | Implementation |
|-------------|----------------|
| API 文档 | 代码注释，godoc 格式 |
| README | 项目根目录 README |
| 架构决策 | research.md 记录所有决策 |

## Project Structure

### Documentation (this feature)

```text
specs/001-notion-mcp-tools/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
notion-mcp/
├── cmd/
│   ├── root.go          # Cobra root command
│   └── serve.go         # serve subcommand
├── internal/
│   ├── server/
│   │   ├── server.go    # MCP server implementation
│   │   ├── prompts.go   # Prompt handlers
│   │   ├── resources.go # Resource handlers
│   │   └── tools.go     # Tool handlers
│   ├── cache/
│   │   ├── cache.go     # Cache interface
│   │   ├── memory.go    # In-memory cache
│   │   └── file.go      # File system cache
│   ├── notion/
│   │   ├── client.go    # Notion API client
│   │   ├── models.go    # Notion data models
│   │   └── parser.go    # Content block parser
│   ├── tools/
│   │   ├── executor.go  # Code execution
│   │   └── registry.go  # Tool registry
│   └── config/
│       └── config.go    # Configuration loading
├── tests/
│   ├── unit/
│   ├── integration/
│   └── contract/
├── .env.example
├── go.mod
├── go.sum
├── main.go
└── README.md
```

**Structure Decision**: 标准 Go 项目布局，cmd/ 存放入口，internal/ 存放内部实现，tests/ 存放测试。

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| 双层缓存 | 性能需求：内存缓存减少文件 IO，文件缓存持久化 | 单层文件缓存：重启后需要重新加载 Notion |
| 代码执行 | 用户需求：将 Notion 代码块作为工具 | 仅读取代码块：无法动态执行工具功能 |

## Phase 0: Research Complete ✓

Research findings documented in `research.md`:
- MCP Go SDK 使用模式
- Notion API 端点和数据模型
- 代码执行策略
- 缓存架构
- 速率限制处理

## Phase 1: Design Artifacts

**Prerequisites**: research.md complete

### Deliverables

- [ ] `data-model.md` - 数据模型定义
- [ ] `contracts/mcp-tools.json` - 工具 Schema
- [ ] `contracts/mcp-resources.json` - 资源 Schema
- [ ] `contracts/mcp-prompts.json` - 提示词 Schema
- [ ] `quickstart.md` - 快速开始指南
- [ ] Agent context update (CLAUDE.md)

### Next Steps

Execute `/speckit.tasks` to generate implementation tasks based on:
- User stories from spec.md
- Technical design from plan.md
- Data models from data-model.md
