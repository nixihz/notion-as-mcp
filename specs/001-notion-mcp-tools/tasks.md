---

description: "Task list for Notion MCP Server implementation"
---

# Tasks: Notion MCP Server

**Input**: Design documents from `/specs/001-notion-mcp-tools/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), data-model.md, contracts/, research.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Source**: `cmd/`, `internal/` at repository root
- **Tests**: `tests/` at repository root
- Paths shown below assume single project - adjust based on plan.md structure

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Create project structure per implementation plan
- [x] T002 Initialize Go 1.25 project with go.mod
- [x] T003 [P] Configure linting (golangci-lint) and formatting (gofmt)
- [x] T004 [P] Create .env.example with all configuration variables
- [x] T005 [P] Create main.go entry point

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

### Configuration

- [x] T006 [P] Implement config loading in internal/config/config.go
- [x] T007 [P] Add environment variable parsing with godotenv
- [x] T008 [P] Define Config struct matching data-model.md

### Notion Client (Foundation for US1, US2, US3, US6)

- [x] T009 [P] Implement Notion API client in internal/notion/client.go
- [x] T010 [P] Add HTTP client with retry logic (exponential backoff)
- [x] T011 [P] Implement database query endpoint
- [x] T012 [P] Implement page content retrieval (blocks endpoint)
- [x] T013 Define Notion data models in internal/notion/models.go

### Cache Infrastructure (Foundation for US4)

- [x] T014 Define Cache interface in internal/cache/cache.go
- [x] T015 [P] Implement in-memory cache in internal/cache/memory.go
- [x] T016 [P] Implement file system cache in internal/cache/file.go

### Logging (Foundation for US5)

- [x] T017 Setup slog logger in internal/logger/logger.go
- [x] T018 [P] Add log level configuration support

### MCP Server Foundation

- [x] T019 Initialize MCP server in internal/server/server.go
- [x] T020 [P] Setup stdio transport connection

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Notion Database Integration (Priority: P1) 🎯 MVP

**Goal**: 开发者配置并启动 MCP 服务器，服务器成功连接到 Notion 数据库并能够读取其中的条目

**Independent Test**: 启动服务器后，配置有效的 Notion API 密钥和数据库 ID，服务器成功连接并返回数据库中的条目列表

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T021 [P] [US1] Contract test for Notion client in tests/contract/test_notion_client.go
- [ ] T022 [P] [US1] Integration test for server initialization in tests/integration/test_server_init.go

### Implementation for User Story 1

- [x] T023 [P] [US1] Create Cobra root command in cmd/root.go
- [x] T024 [P] [US1] Create serve subcommand in cmd/serve.go
- [x] T025 [US1] Implement server startup sequence (depends on T019, T006)
- [x] T026 [US1] Add Notion connection validation on startup
- [x] T027 [US1] Add error handling for invalid credentials
- [x] T028 [US1] Add logging for connection status

**Checkpoint**: User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Prompt and Resource Extraction (Priority: P1)

**Goal**: 从 Notion 数据库中提取 prompt 和 resource 类型的条目内容，供 MCP 客户端使用

**Independent Test**: 在 Notion 数据库中添加类型为 "prompt" 或 "resource" 的条目，客户端能够成功读取其文本内容

### Tests for User Story 2

- [ ] T029 [P] [US2] Unit test for content parser in tests/unit/test_parser.go
- [ ] T030 [P] [US2] Integration test for prompt extraction in tests/integration/test_prompts.go

### Implementation for User Story 2

- [x] T031 [P] [US2] Implement content block parser in internal/notion/parser.go
- [x] T032 [P] [US2] Create Prompt struct in internal/server/prompts.go
- [x] T033 [P] [US2] Create Resource struct in internal/server/resources.go
- [x] T034 [US2] Implement MCP prompts list handler (depends on T031)
- [x] T035 [US2] Implement MCP prompts get handler (depends on T034)
- [x] T036 [US2] Implement MCP resources list handler (depends on T031)
- [x] T037 [US2] Implement MCP resources read handler (depends on T036)
- [x] T038 [US2] Handle nested content and text flattening

**Checkpoint**: User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Tool Code Block Extraction (Priority: P1)

**Goal**: 从 Notion 数据库的 tool 类型条目中提取代码块，作为可执行工具暴露给 MCP 客户端

**Independent Test**: 在 Notion 数据库中添加类型为 "tool" 的条目，其中包含代码块，客户端能够将该代码块作为工具调用

### Tests for User Story 3

- [ ] T039 [P] [US3] Unit test for tool registry in tests/unit/test_registry.go
- [ ] T040 [P] [US3] Integration test for tool execution in tests/integration/test_tools.go

### Implementation for User Story 3

- [x] T041 [P] [US3] Implement tool registry in internal/tools/registry.go
- [x] T042 [P] [US3] Implement code executor in internal/tools/executor.go
- [x] T043 [P] [US3] Parse code block language and content
- [x] T044 [US3] Implement MCP tools list handler (depends on T041, T031)
- [x] T045 [US3] Implement MCP tools call handler (depends on T042, T044)
- [x] T046 [US3] Add timeout handling for code execution
- [x] T047 [US3] Add error formatting for tool execution results
- [x] T048 [US3] Add support for bash/python execution

**Checkpoint**: All three P1 user stories should be independently functional

---

## Phase 6: User Story 6 - Type-based Filtering (Priority: P1)

**Goal**: 根据 Notion 条目中的类型字段，过滤和分类不同的 MCP 原语

**Independent Test**: 在 Notion 数据库中有多种类型的条目，客户端只能看到对应类型的条目

### Implementation for User Story 6

- [ ] T049 [P] [US6] Implement type filter in internal/notion/client.go
- [ ] T050 [P] [US6] Add type mapping configuration
- [ ] T051 [US6] Integrate type filter with database queries
- [ ] T052 [US6] Handle unknown/invalid type values
- [ ] T053 [US6] Add multi-type support per entry

**Note**: US6 is integrated with US2 and US3, not a separate checkpoint

---

## Phase 7: User Story 4 - Caching Layer (Priority: P2)

**Goal**: 实现本地缓存机制，减少对 Notion API 的实时请求频率

**Independent Test**: 连续两次请求相同数据，第二次请求应该从缓存返回，响应时间显著降低

### Implementation for User Story 4

- [ ] T054 [P] [US4] Implement cache key generation
- [ ] T055 [P] [US4] Add TTL validation logic
- [ ] T056 [US4] Integrate cache with Notion client
- [ ] T057 [US4] Add cache invalidation on data change
- [ ] T058 [US4] Handle cache write failures gracefully
- [ ] T059 [US4] Add cache statistics logging

---

## Phase 8: User Story 5 - Logging and Observability (Priority: P2)

**Goal**: 提供直观的日志输出，使开发者能够了解服务器运行状态

**Independent Test**: 启动服务器并执行操作，日志中显示连接、请求、执行结果等关键信息

### Implementation for User Story 5

- [ ] T060 [P] [US5] Add request logging middleware
- [ ] T061 [P] [US5] Add performance timing logging
- [ ] T062 [US5] Implement structured error logging
- [ ] T063 [US5] Add startup configuration summary logging
- [ ] T064 [US5] Implement log level configuration

---

## Phase 9: Change Detection (Polling)

**Purpose**: 自动检测 Notion 数据变更

- [ ] T065 [P] Implement change detector in internal/notion/changedetector.go
- [ ] T066 [P] Add last_edited_time tracking
- [ ] T067 Implement polling worker with configurable interval
- [ ] T068 Add REFRESH_ON_START functionality
- [ ] T069 Integrate change detection with cache invalidation

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T070 [P] Add README.md documentation
- [ ] T071 [P] Add godoc comments to all public APIs
- [ ] T072 Run full test suite and ensure >80% coverage
- [ ] T073 [P] Performance optimization for large databases
- [ ] T074 Add rate limit handling verification
- [ ] T075 Final integration testing with real Notion database

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-8)**: All depend on Foundational phase completion
  - US1, US2, US3, US6 can proceed in parallel (after Phase 2)
  - US4, US5 depend on their respective foundations being complete
- **Change Detection (Phase 9)**: Depends on Cache and Notion Client
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **US1 (Database Integration)**: Can start after Foundational - No dependencies on other stories
- **US2 (Prompt/Resource)**: Can start after Foundational - Depends on US1 Notion client
- **US3 (Tools)**: Can start after Foundational - Depends on US1 Notion client, US6 type filtering
- **US4 (Caching)**: Can start after Foundational - Can proceed in parallel
- **US5 (Logging)**: Can start after Foundational - Can proceed in parallel
- **US6 (Type Filtering)**: Can start after Foundational - Depends on US1 Notion client

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Models before services
- Services before endpoints
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all P1 user stories can start in parallel
- All tests for a user story marked [P] can run in parallel
- Models within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Contract test for Notion client in tests/contract/test_notion_client.go"
Task: "Integration test for server initialization in tests/integration/test_server_init.go"

# Launch all models for User Story 1 together:
Task: "Create Cobra root command in cmd/root.go"
Task: "Create serve subcommand in cmd/serve.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 + 3 → Test independently → Deploy/Demo
4. Add User Story 6 → Test → Deploy/Demo
5. Add User Story 4 + 5 → Test → Deploy/Demo
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1
   - Developer B: User Story 2 + 6
   - Developer C: User Story 3
3. Stories complete and integrate independently

---

## Task Summary

| Phase | Task Count | Description |
|-------|------------|-------------|
| Phase 1: Setup | 5 | Project initialization |
| Phase 2: Foundational | 14 | Core infrastructure |
| Phase 3: US1 (Integration) | 8 | Notion connection |
| Phase 4: US2 (Prompt/Resource) | 8 | Content extraction |
| Phase 5: US3 (Tools) | 8 | Code execution |
| Phase 6: US6 (Type Filtering) | 5 | Type classification |
| Phase 7: US4 (Caching) | 6 | Cache layer |
| Phase 8: US5 (Logging) | 5 | Observability |
| Phase 9: Change Detection | 5 | Polling |
| Phase 10: Polish | 6 | Final improvements |
| **Total** | **70** | |

### By User Story

| User Story | Task Count | Priority |
|------------|------------|----------|
| US1: Database Integration | 8 | P1 |
| US2: Prompt/Resource | 8 | P1 |
| US3: Tool Execution | 8 | P1 |
| US4: Caching | 6 | P2 |
| US5: Logging | 5 | P2 |
| US6: Type Filtering | 5 | P1 |
| Shared/Foundation | 30 | - |

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
