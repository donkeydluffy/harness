# 任务列表

- Status: Approved
- Updated: 2026-11-26
- PRD: 1.0-rc3
- Release: 1

Release 1 的任务，按执行顺序编号。字段格式见 `docs/architecture/formats.md` 第 3 节。

约定：
- 集成测试放在 `internal/app/<主题>_integration_test.go`，夹具在 `internal/app/testdata/<主题>/`，测试名包含用例编号（如 `TestCheck_AC_003_1_N1`）。
- 每个提交带 `Task: T-n` 尾注，改动不超过 400 行（不含产物文件和生成文件）。
- 任务的 Covers 可以是 FR、NFR 或 AC（ADR-008）；任务 Done 要求所关联条目涉及的用例全部通过。
- 在对应检查实现之前完成的任务，用 `go test` 与原型脚本验证；检查上线后回溯校验此前全部任务。

## 任务

### T-001 CLI 骨架与 PRD 编号、必填字段检查
- Status: Done
- Covers: AC-003-1, AC-003-2, AC-003-4, NFR-004
- Design: FLOW-01, IF-001, IF-002, IF-009, IF-010, IF-011, IF-012, IF-013, IF-014, IF-022, ADR-001, ADR-002, ADR-007
- Flow: FLOW-01
- Scope: go.mod, cmd/harness/**, internal/cli/**, internal/app/app.go, internal/app/check.go, internal/model/**, internal/workspace/**, internal/rules/engine/**, internal/rules/builtin/registry.go, internal/rules/builtin/registry_test.go, internal/rules/builtin/util.go, internal/rules/builtin/prd_01.go, internal/rules/builtin/testdata/prd.01/**, internal/rules/builtin/prd_02.go, internal/rules/builtin/testdata/prd.02/**, internal/app/prd_integration_test.go, internal/app/testdata/prd/**
- Depends: 

端到端骨架（prd.02 同时实现 Release、Current-Release 与 NFR Type 字段校验，AC-003-5 的追溯部分在 T-006 验收）：model 解析器、workspace 加载配置与产物、engine 注册与运行（含样本自检与修复建议必填）、app.Check、`harness check [--json]` 与退出码 0/1/2。完成后 `harness check prd.01 prd.02` 可检查本仓库。engine 附 `TestExample_EXT_03` 示例规则测试。

### T-002 锁文件、`harness lock` 与需求变更追踪
- Status: Done
- Covers: FR-004, AC-003-3, AC-019-3
- Design: FLOW-10, IF-005, IF-006, IF-023, ADR-004
- Flow: FLOW-01
- Scope: internal/lock/**, internal/workspace/**, internal/cli/**, internal/app/lock.go, internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/prd_03.go, internal/rules/builtin/testdata/prd.03/**, internal/rules/builtin/prd_08.go, internal/rules/builtin/testdata/prd.08/**, internal/app/lock_integration_test.go, internal/app/testdata/lock/**
- Depends: T-001

规范化与哈希、锁文件读写（键排序、逐字节稳定）、`harness lock`（检查通过才写，保留批准与阶段）、prd.03 编号不得消失或复用、prd.08 CHG 与哈希差异一致。

### T-003 澄清记录、验收用例与需求文本质量
- Status: Done
- Covers: FR-002, AC-019-1, AC-019-2, FR-040
- Design: FLOW-01, IF-014
- Flow: FLOW-01
- Scope: internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/prd_04.go, internal/rules/builtin/testdata/prd.04/**, internal/rules/builtin/prd_05.go, internal/rules/builtin/testdata/prd.05/**, internal/rules/builtin/prd_06.go, internal/rules/builtin/testdata/prd.06/**, internal/rules/builtin/prd_07.go, internal/rules/builtin/testdata/prd.07/**, internal/rules/builtin/prd_09.go, internal/rules/builtin/testdata/prd.09/**, internal/rules/builtin/prd_10.go, internal/rules/builtin/testdata/prd.10/**, internal/app/clarify_integration_test.go, internal/app/testdata/clarify/**
- Depends: T-002

三章齐全、Open 待澄清项阻止 Approved、澄清记录按轮次且每轮不超过 5 问、用例编号与 N 加 B/E、模糊词（跳过代码格式）、Quantitative NFR 含数值与单位。

### T-004 统一语言一致性
- Status: Done
- Covers: FR-005
- Design: FLOW-01, IF-014
- Flow: FLOW-01
- Scope: internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/term_01.go, internal/rules/builtin/testdata/term.01/**, internal/rules/builtin/term_02.go, internal/rules/builtin/testdata/term.02/**, internal/app/term_integration_test.go, internal/app/testdata/term/**
- Depends: T-003

### T-005 架构产物校验
- Status: Done
- Covers: FR-006, AC-033-2
- Design: FLOW-01, MOD-03, IF-002, IF-014
- Flow: FLOW-01
- Scope: internal/model/**, internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/arch_01.go, internal/rules/builtin/testdata/arch.01/**, internal/rules/builtin/arch_02.go, internal/rules/builtin/testdata/arch.02/**, internal/rules/builtin/arch_03.go, internal/rules/builtin/testdata/arch.03/**, internal/app/arch_integration_test.go, internal/app/testdata/arch/**
- Depends: T-004

模块、接口、流程、扩展点必填字段与模块路径不重叠，依赖无环，流程步骤与依赖一致。

### T-006 需求到设计追溯与覆盖矩阵
- Status: Done
- Covers: FR-007, AC-003-5
- Design: FLOW-01, IF-014, IF-022
- Flow: FLOW-01
- Scope: internal/cli/**, internal/app/trace.go, internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/trace_01.go, internal/rules/builtin/testdata/trace.01/**, internal/rules/builtin/trace_02.go, internal/rules/builtin/testdata/trace.02/**, internal/rules/builtin/trace_03.go, internal/rules/builtin/testdata/trace.03/**, internal/app/trace_integration_test.go, internal/app/testdata/trace/**
- Depends: T-005

trace.01 ~ trace.03 与 `harness trace --format md|json`。完成后退役原型脚本 selfcheck.py。

### T-007 git 访问与架构决策记录检查
- Status: Done
- Covers: FR-010
- Design: FLOW-01, IF-007, IF-009, IF-014
- Flow: FLOW-01
- Scope: internal/vcs/**, internal/workspace/**, internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/adr_01.go, internal/rules/builtin/testdata/adr.01/**, internal/rules/builtin/adr_02.go, internal/rules/builtin/testdata/adr.02/**, internal/rules/builtin/adr_03.go, internal/rules/builtin/testdata/adr.03/**, internal/app/adr_integration_test.go, internal/app/testdata/adr/**
- Depends: T-006

vcs 通过 `git` 命令实现提交范围、改动、指定版本文件；adr.03 判断模块增删、依赖变化、扩展点变化是否在同一提交中被 ADR 引用。

### T-008 测试报告解析、Go 语言包与集成测试追溯
- Status: Done
- Covers: FR-020
- Design: FLOW-01, IF-003, IF-004, IF-008, IF-009, IF-014
- Flow: FLOW-01
- Scope: internal/testreport/**, internal/lang/**, internal/workspace/**, internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/test_01.go, internal/rules/builtin/testdata/test.01/**, internal/rules/builtin/test_02.go, internal/rules/builtin/testdata/test.02/**, internal/app/tests_integration_test.go, internal/app/testdata/tests/**
- Depends: T-007

JUnit XML 与 `go test -json` 解析；Go 语言包（注释提取、`*_integration_test.go` 判定为集成测试）；test.01、test.02。lang 附 `TestExample_EXT_01` 示例语言测试。完成后本仓库的测试报告写入 `.harness/reports/`。

### T-009 技术债登记检查
- Status: Done
- Covers: FR-031
- Design: FLOW-01, IF-003, IF-014
- Flow: FLOW-01
- Scope: internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/debt_01.go, internal/rules/builtin/testdata/debt.01/**, internal/rules/builtin/debt_02.go, internal/rules/builtin/testdata/debt.02/**, internal/rules/builtin/debt_03.go, internal/rules/builtin/testdata/debt.03/**, internal/app/debt_integration_test.go, internal/app/testdata/debt/**
- Depends: T-008

### T-010 扩展场景与扩展点规则
- Status: InProgress
- Covers: AC-033-1, AC-033-3, AC-033-4
- Design: FLOW-01, IF-014
- Flow: FLOW-01
- Scope: internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/prd_11.go, internal/rules/builtin/testdata/prd.11/**, internal/rules/builtin/arch_04.go, internal/rules/builtin/testdata/arch.04/**, internal/rules/builtin/arch_05.go, internal/rules/builtin/testdata/arch.05/**, internal/app/ext_integration_test.go, internal/app/testdata/ext/**
- Depends: T-009

### T-011 任务列表检查
- Status: Todo
- Covers: AC-011-1, AC-011-2, AC-011-3, AC-011-4, AC-011-6, AC-034-2
- Design: FLOW-01, IF-014, ADR-008
- Flow: FLOW-01
- Scope: internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/task_01.go, internal/rules/builtin/testdata/task.01/**, internal/rules/builtin/task_02.go, internal/rules/builtin/testdata/task.02/**, internal/rules/builtin/task_03.go, internal/rules/builtin/testdata/task.03/**, internal/rules/builtin/task_04.go, internal/rules/builtin/testdata/task.04/**, internal/rules/builtin/task_05.go, internal/rules/builtin/testdata/task.05/**, internal/rules/builtin/task_06.go, internal/rules/builtin/testdata/task.06/**, internal/app/task_integration_test.go, internal/app/testdata/task/**
- Depends: T-010

task.02 按 ADR-008 判定覆盖：FR 被直接关联，或其每条 Active AC 都被关联。完成后本文件由 `harness check task` 校验。

### T-012 提交检查
- Status: Todo
- Covers: AC-013-1, AC-013-2, AC-013-4, AC-034-1
- Design: FLOW-09, IF-007, IF-014, IF-022, ADR-008
- Flow: FLOW-01
- Scope: internal/cli/**, internal/app/check.go, internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/commit_01.go, internal/rules/builtin/testdata/commit.01/**, internal/rules/builtin/commit_02.go, internal/rules/builtin/testdata/commit.02/**, internal/rules/builtin/commit_03.go, internal/rules/builtin/testdata/commit.03/**, internal/rules/builtin/commit_04.go, internal/rules/builtin/testdata/commit.04/**, internal/app/commit_integration_test.go, internal/app/testdata/commit/**
- Depends: T-011

`harness check commits <范围>`；commit.02 豁免产物文件和生成文件。

### T-013 人工批准与批准有效性
- Status: Todo
- Covers: AC-037-1, AC-037-3
- Design: FLOW-08, IF-005, IF-006, IF-024
- Flow: FLOW-01
- Scope: internal/cli/**, internal/app/approve.go, internal/lock/**, internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/approve_01.go, internal/rules/builtin/testdata/approve.01/**, internal/app/approve_integration_test.go, internal/app/testdata/approve/**
- Depends: T-012

`harness approve <产物>` 要求交互式终端并确认，写入批准人、时间、哈希；approve.01 内容与批准哈希不一致时视为 Draft。完成后用户在终端补录 PRD 与架构的批准。

### T-014 流程阶段定义与推进
- Status: Todo
- Covers: AC-045-1, AC-045-2, AC-045-3
- Design: FLOW-08, IF-015, IF-016, IF-024, ADR-005
- Flow: FLOW-08
- Scope: internal/phase/**, internal/lock/**, internal/cli/**, internal/app/phase.go, internal/app/app.go, internal/app/phase_integration_test.go, internal/app/testdata/phase/**
- Depends: T-013

内置四个阶段 JSON、项目配置合并、phase.01（由 phase 模块注册）、`harness phase next`（交互式终端、退出与进入规则通过）。phase 附 `TestExample_EXT_04` 示例阶段测试。

### T-015 阶段前置条件与任务完成定义
- Status: Todo
- Covers: FR-038, FR-039, AC-024-2
- Design: FLOW-01, IF-014, ADR-008
- Flow: FLOW-01
- Scope: internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/task_07.go, internal/rules/builtin/testdata/task.07/**, internal/rules/builtin/task_08.go, internal/rules/builtin/testdata/task.08/**, internal/app/dod_integration_test.go, internal/app/testdata/dod/**
- Depends: T-014

task.07：InProgress 任务要求 PRD 与所引用设计的批准有效、关联需求无 Open 待澄清项。task.08：Done 任务所关联条目涉及的用例全部 passed、全量检查通过、进度文件有记录。

### T-016 写操作门禁与 Claude Code 适配器
- Status: Todo
- Covers: AC-012-1, AC-012-2, AC-012-5, AC-037-2, AC-045-4
- Design: FLOW-02, FLOW-03, IF-017, IF-018, IF-019, IF-025, ADR-003, ADR-006
- Flow: FLOW-02
- Scope: go.mod, go.sum, internal/gate/**, internal/adapter/**, internal/cli/**, internal/app/gate.go, internal/app/gate_integration_test.go, internal/app/testdata/gate/**
- Depends: T-015

gate.protect、gate.artifact、gate.write；Bash 命令用 mvdan.cc/sh 事前解析，事后快照比对与回滚（大于 1 MiB 只报告）；claudecode 适配器解码 PreToolUse、编码 deny，无法解析即拒绝；`harness gate`。adapter 附 `TestExample_EXT_02` 示例适配器测试。

### T-017 安全边界
- Status: Todo
- Covers: FR-025
- Design: FLOW-02, IF-017
- Flow: FLOW-02
- Scope: internal/gate/**, internal/app/security_integration_test.go, internal/app/testdata/security/**
- Depends: T-016

gate.sensitive、gate.secret、gate.command；敏感路径、密钥规则、危险命令可配置。

### T-018 写后检查、会话结束检查与会话开始注入
- Status: Todo
- Covers: AC-012-3, AC-012-4, AC-024-1, AC-024-3
- Design: FLOW-04, FLOW-05, FLOW-06, IF-012, IF-017, IF-019, IF-025
- Flow: FLOW-05
- Scope: internal/gate/**, internal/adapter/**, internal/cli/**, internal/app/gate.go, internal/app/session_integration_test.go, internal/app/testdata/session/**
- Depends: T-017

PostToolUse 写产物后按类型运行规则并反馈；Stop 全量检查，连续阻止 3 次后放行并写进度文件；SessionStart 注入阶段、InProgress 任务、未通过检查、Open 待澄清项、冒烟测试结果。

### T-019 项目初始化与产物模板
- Status: Todo
- Covers: FR-001, AC-N002-2, NFR-005
- Design: FLOW-07, IF-020, IF-021, IF-026
- Flow: FLOW-07
- Scope: internal/scaffold/**, internal/cli/**, internal/app/init.go, internal/app/init_integration_test.go, internal/app/testdata/init/**
- Depends: T-018

go:embed 模板；`harness init --lang go` 生成产物骨架、harness.json、Claude Code hook 配置；重复执行不覆盖已修改文件并列出跳过的文件。

### T-020 流程 skill 集
- Status: Todo
- Covers: FR-016, AC-011-5, AC-033-5
- Design: MOD-13, IF-020
- Scope: internal/scaffold/templates/skills/**, internal/scaffold/skills_integration_test.go
- Depends: T-019

clarify、change、architect、plan、implement、test 六个 skill；每个声明结束前运行的 `harness check` 子命令；静态测试校验声明存在、测试命名约定与 test.01 的提取规则一致。

### T-021 git hook、CI 模板与生成文件一致性
- Status: Todo
- Covers: AC-013-3, AC-043-1, AC-041-2
- Design: FLOW-07, IF-020, IF-021
- Flow: FLOW-07
- Scope: internal/scaffold/**, internal/app/init.go, internal/app/ci_integration_test.go, internal/app/testdata/ci/**
- Depends: T-020

commit-msg、pre-commit hook；GitHub Actions 与 GitLab CI 模板运行 `harness check` 和 `harness check commits`；cfg.02（由 scaffold 模块注册）。

### T-022 门禁配置锁定与 CI 检查步骤
- Status: Todo
- Covers: AC-041-1, AC-043-2
- Design: FLOW-01, IF-014
- Flow: FLOW-01
- Scope: internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/cfg_01.go, internal/rules/builtin/testdata/cfg.01/**, internal/rules/builtin/ci_01.go, internal/rules/builtin/testdata/ci.01/**, internal/app/cfg_integration_test.go, internal/app/testdata/cfg/**
- Depends: T-021

### T-023 扩展场景回归：新增阶段与新增规则
- Status: Todo
- Covers: NFR-008, NFR-009
- Design: FLOW-09, EXT-03, EXT-04
- Scope: internal/app/extdemo_integration_test.go, internal/app/testdata/extdemo/**
- Depends: T-022

在临时 git 仓库中按 EXT-04、EXT-03 各做一次示例扩展提交，验证只改允许路径时 commit.04 通过、改动封闭模块时失败。

### T-024 确定性与无网络访问
- Status: Todo
- Covers: NFR-001
- Design: FLOW-01, ADR-001
- Scope: internal/app/determinism_integration_test.go, internal/app/testdata/determinism/**
- Depends: T-023

静态测试：除 audit 包外不得引入网络包；全部检查子命令在禁用网络时结果一致。

### T-025 性能基准
- Status: Todo
- Covers: NFR-003
- Design: FLOW-01, FLOW-02
- Scope: internal/app/perf_integration_test.go, internal/app/testdata/perf/**
- Depends: T-024

生成 5000 个源文件、200 个任务的样例仓库；PreToolUse 门禁 p95 < 200ms，全量检查 < 10s，50000 行 PRD 的 `harness check prd` < 2s。

### T-026 自举：本仓库启用 harness
- Status: Todo
- Covers: FR-017
- Design: FLOW-01, FLOW-07
- Scope: internal/app/bootstrap_integration_test.go, internal/app/testdata/bootstrap/**
- Depends: T-025

对本仓库执行 `harness init`，提交生成文件与锁文件，启用 hook 与 CI；集成测试对仓库根目录运行 `harness check` 并校验 hook 配置存在。

### T-027 跨平台构建与发布
- Status: Todo
- Covers: AC-N002-1, AC-N002-3
- Design: MOD-01, ADR-001
- Scope: cmd/harness/**, .goreleaser.yaml, .github/workflows/release.yml
- Depends: T-026

6 个平台交叉编译、GitHub Release 附校验和、Homebrew tap formula。需要 GitHub 仓库与 tap 仓库，发布前由用户确认。
