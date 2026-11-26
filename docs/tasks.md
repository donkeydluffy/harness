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
- Status: Done
- Covers: AC-033-1, AC-033-3, AC-033-4
- Design: FLOW-01, IF-014
- Flow: FLOW-01
- Scope: internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/prd_11.go, internal/rules/builtin/testdata/prd.11/**, internal/rules/builtin/arch_04.go, internal/rules/builtin/testdata/arch.04/**, internal/rules/builtin/arch_05.go, internal/rules/builtin/testdata/arch.05/**, internal/app/ext_integration_test.go, internal/app/testdata/ext/**
- Depends: T-009

### T-011 任务列表检查
- Status: Done
- Covers: AC-011-1, AC-011-2, AC-011-3, AC-011-4, AC-011-6, AC-034-2
- Design: FLOW-01, IF-014, ADR-008
- Flow: FLOW-01
- Scope: internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/task_01.go, internal/rules/builtin/testdata/task.01/**, internal/rules/builtin/task_02.go, internal/rules/builtin/testdata/task.02/**, internal/rules/builtin/task_03.go, internal/rules/builtin/testdata/task.03/**, internal/rules/builtin/task_04.go, internal/rules/builtin/testdata/task.04/**, internal/rules/builtin/task_05.go, internal/rules/builtin/testdata/task.05/**, internal/rules/builtin/task_06.go, internal/rules/builtin/testdata/task.06/**, internal/app/task_integration_test.go, internal/app/testdata/task/**
- Depends: T-010

task.02 按 ADR-008 判定覆盖：FR 被直接关联，或其每条 Active AC 都被关联。完成后本文件由 `harness check task` 校验。

### T-012 提交检查
- Status: Done
- Covers: AC-013-1, AC-013-2, AC-013-4, AC-034-1
- Design: FLOW-09, IF-007, IF-014, IF-022, ADR-008
- Flow: FLOW-01
- Scope: internal/cli/**, internal/app/check.go, internal/workspace/workspace.go, internal/rules/builtin/registry.go, internal/rules/builtin/registry_test.go, internal/rules/builtin/util.go, internal/rules/builtin/commit_01.go, internal/rules/builtin/testdata/commit.01/**, internal/rules/builtin/commit_02.go, internal/rules/builtin/testdata/commit.02/**, internal/rules/builtin/commit_03.go, internal/rules/builtin/testdata/commit.03/**, internal/rules/builtin/commit_04.go, internal/rules/builtin/testdata/commit.04/**, internal/app/commit_integration_test.go, internal/app/testdata/commit/**
- Depends: T-011

`harness check commits <范围>`；commit.02 豁免产物文件和生成文件。

### T-013 人工批准与批准有效性
- Status: Done
- Covers: AC-037-1, AC-037-3
- Design: FLOW-08, IF-005, IF-006, IF-024
- Flow: FLOW-01
- Scope: internal/cli/**, internal/app/approve.go, internal/workspace/approval.go, internal/lock/**, internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/approve_01.go, internal/rules/builtin/testdata/approve.01/**, internal/app/approve_integration_test.go, internal/app/testdata/approve/**
- Depends: T-012

`harness approve <产物>` 要求交互式终端并确认，写入批准人、时间、哈希；approve.01 内容与批准哈希不一致时视为 Draft。完成后用户在终端补录 PRD 与架构的批准。

### T-014 流程阶段定义与推进
- Status: Done
- Covers: AC-045-1, AC-045-2
- Design: FLOW-08, IF-015, IF-016, IF-024, ADR-005
- Flow: FLOW-08
- Scope: internal/phase/**, internal/lock/**, internal/cli/**, internal/app/phase.go, internal/app/app.go, internal/app/approve.go, internal/app/phase_integration_test.go, internal/app/testdata/phase/**
- Depends: T-013

内置四个阶段 JSON、项目配置合并、phase.01（由 phase 模块注册）、`harness phase list|next`（交互式终端、退出与进入规则通过）。AC-045-3 的 `phase next` 在本任务实现并写测试，但其 E3（门禁拒绝代理改阶段）依赖门禁，整条 AC 由 T-016 关联。phase 附 `TestExample_EXT_04` 示例阶段测试。

### T-015 阶段前置条件与任务完成定义
- Status: Done
- Covers: FR-038, FR-039, AC-024-2
- Design: FLOW-01, IF-014, ADR-008
- Flow: FLOW-01
- Scope: internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/test_02.go, internal/app/task_integration_test.go, internal/rules/builtin/task_07.go, internal/rules/builtin/testdata/task.07/**, internal/rules/builtin/task_08.go, internal/rules/builtin/testdata/task.08/**, internal/app/dod_integration_test.go, internal/app/testdata/dod/**
- Depends: T-014

task.07：InProgress 任务要求 PRD 与所引用设计的批准有效、关联需求无 Open 待澄清项。task.08：Done 任务所关联条目涉及的用例全部 passed、全量检查通过、进度文件有记录。

### T-016 写操作门禁与 Claude Code 适配器
- Status: Done
- Covers: AC-012-1, AC-012-2, AC-012-5, AC-037-2, AC-045-3, AC-045-4
- Design: FLOW-02, FLOW-03, IF-017, IF-018, IF-019, IF-025, ADR-003, ADR-006
- Flow: FLOW-02
- Scope: go.mod, go.sum, internal/gate/**, internal/adapter/**, internal/cli/**, internal/app/gate.go, internal/app/gate_integration_test.go, internal/app/testdata/gate/**
- Depends: T-015

gate.protect、gate.artifact、gate.write；Bash 命令用 mvdan.cc/sh 事前解析，事后快照比对与回滚（大于 1 MiB 只报告）；claudecode 适配器解码 PreToolUse、编码 deny，无法解析即拒绝；`harness gate`。adapter 附 `TestExample_EXT_02` 示例适配器测试。

### T-017 安全边界
- Status: Done
- Covers: FR-025
- Design: FLOW-02, IF-017
- Flow: FLOW-02
- Scope: internal/gate/**, internal/app/gate.go, internal/app/security_integration_test.go, internal/app/testdata/security/**
- Depends: T-016

gate.sensitive、gate.secret、gate.command；敏感路径、密钥规则、危险命令可配置。

### T-018 写后检查、会话结束检查与会话开始注入
- Status: Done
- Covers: AC-012-3, AC-012-4, AC-024-1, AC-024-3
- Design: FLOW-04, FLOW-05, FLOW-06, IF-012, IF-017, IF-019, IF-025
- Flow: FLOW-05
- Scope: internal/gate/**, internal/adapter/**, internal/cli/**, internal/rules/engine/engine.go, internal/app/gate.go, internal/app/session_integration_test.go, internal/app/testdata/session/**
- Depends: T-017

PostToolUse 写产物后按类型运行规则并反馈；Stop 全量检查，连续阻止 3 次后放行并写进度文件；SessionStart 注入阶段、InProgress 任务、未通过检查、Open 待澄清项、冒烟测试结果。

### T-019 项目初始化与产物模板
- Status: Done
- Covers: FR-001, AC-N002-2, NFR-005
- Design: FLOW-07, IF-020, IF-021, IF-026
- Flow: FLOW-07
- Scope: internal/scaffold/**, internal/lang/**, internal/cli/**, internal/app/init.go, internal/app/init_integration_test.go, internal/app/testdata/init/**
- Depends: T-018

go:embed 模板；`harness init --lang go` 生成产物骨架、harness.json、Claude Code hook 配置；重复执行不覆盖已修改文件并列出跳过的文件。

### T-020 流程 skill 集
- Status: Done
- Covers: FR-016, AC-011-5, AC-033-5
- Design: MOD-13, IF-020, IF-026, FLOW-07
- Flow: FLOW-07
- Scope: internal/scaffold/templates/skills/**, internal/scaffold/scaffold.go, internal/scaffold/skills_integration_test.go, internal/app/init_integration_test.go
- Depends: T-019

clarify、change、architect、plan、implement、test 六个 skill；每个声明结束前运行的 `harness check` 子命令；静态测试校验声明存在、测试命名约定与 test.01 的提取规则一致。

### T-021 git hook、CI 模板与生成文件一致性
- Status: Done
- Covers: AC-013-3, AC-043-1, AC-041-2
- Design: FLOW-07, IF-020, IF-021
- Flow: FLOW-07
- Scope: internal/scaffold/**, internal/app/init.go, internal/app/app.go, internal/app/init_integration_test.go, internal/app/session_integration_test.go, internal/cli/init.go, internal/cli/cli.go, internal/app/ci_integration_test.go, internal/app/testdata/ci/**
- Depends: T-020

commit-msg、pre-commit hook；GitHub Actions 与 GitLab CI 模板运行 `harness check` 和 `harness check commits`；cfg.02（由 scaffold 模块注册）。

### T-022 门禁配置锁定与 CI 检查步骤
- Status: Done
- Covers: AC-041-1, AC-043-2
- Design: FLOW-01, IF-014
- Flow: FLOW-01
- Scope: internal/workspace/lockitems.go, internal/rules/builtin/prd_08.go, internal/app/session_integration_test.go, internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/cfg_01.go, internal/rules/builtin/testdata/cfg.01/**, internal/rules/builtin/ci_01.go, internal/rules/builtin/testdata/ci.01/**, internal/app/cfg_integration_test.go, internal/app/testdata/cfg/**
- Depends: T-021

### T-023 扩展场景回归：新增阶段与新增规则
- Status: Done
- Covers: NFR-008, NFR-009
- Design: FLOW-09, EXT-03, EXT-04
- Scope: internal/app/extdemo_integration_test.go, internal/app/testdata/extdemo/**
- Depends: T-022

在临时 git 仓库中按 EXT-04、EXT-03 各做一次示例扩展提交，验证只改允许路径时 commit.04 通过、改动封闭模块时失败。

### T-024 确定性与无网络访问
- Status: Done
- Covers: NFR-001
- Design: FLOW-01, ADR-001
- Scope: internal/app/determinism_integration_test.go, internal/app/testdata/determinism/**
- Depends: T-023

静态测试：除 audit 包外不得引入网络包；全部检查子命令在禁用网络时结果一致。

### T-025 性能基准
- Status: Done
- Covers: NFR-003
- Design: FLOW-01, FLOW-02
- Scope: internal/app/perf_integration_test.go, internal/app/testdata/perf/**
- Depends: T-024

生成 5000 个源文件、200 个任务的样例仓库；PreToolUse 门禁 p95 < 200ms，全量检查 < 10s，50000 行 PRD 的 `harness check prd` < 2s。

### T-026 自举：本仓库启用 harness
- Status: Done
- Covers: FR-017
- Design: FLOW-01, FLOW-07
- Scope: internal/rules/builtin/util.go, internal/app/bootstrap_integration_test.go, internal/app/testdata/bootstrap/**
- Depends: T-025

对本仓库执行 `harness init`，提交生成文件与锁文件，启用 hook 与 CI；集成测试对仓库根目录运行 `harness check` 并校验 hook 配置存在。

### T-027 跨平台构建与发布
- Status: Done
- Covers: AC-N002-1, AC-N002-3
- Design: MOD-01, ADR-001, FLOW-07
- Scope: cmd/harness/**, .goreleaser.yaml, .github/workflows/release.yml, internal/scaffold/**, internal/app/ci_integration_test.go, internal/app/cfg_integration_test.go
- Depends: T-026

6 个平台交叉编译、GitHub Release 附校验和、Homebrew tap formula。需要 GitHub 仓库与 tap 仓库，发布前由用户确认。

### T-028 门禁修复：纯赋值命令崩溃与路径误判为密钥
- Status: Done
- Covers: AC-012-2, AC-025-2
- Design: FLOW-02
- Flow: FLOW-02
- Scope: internal/gate/**, internal/app/gate_integration_test.go, internal/app/security_integration_test.go
- Depends: T-026

T-027 期间发现：只含变量赋值的命令（如 `d=$(mktemp -d)`）使门禁 panic 并以退出码 2 拒绝命令；secret.entropy 把含大小写与数字的长路径当作密钥（DEBT-006）。T-027 在本任务完成前为 Blocked。

### T-029 解析测试不再断言本仓库产物的固定数量
- Status: Done
- Covers: AC-003-1
- Design: IF-001
- Scope: internal/model/model_test.go
- Depends: T-028

`TestParseRepoArtifacts_IF_001` 用固定数字断言本仓库的需求、模块、接口、任务数量，新增 T-028 后失败。改为下限断言，测试只保证本仓库产物能无诊断地解析。

### T-030 找不到 harness 二进制时 hook 失败即阻断（DEBT-004）
- Status: Done
- Covers: AC-012-5
- Design: IF-019, FLOW-02
- Flow: FLOW-02
- Scope: internal/adapter/**, internal/app/hookcmd_integration_test.go
- Depends: T-029

生成的 hook 命令先检查 harness 是否可执行：PreToolUse 找不到时输出拒绝，SessionStart 注入"harness 未安装"，PostToolUse 与 Stop 不阻止。

### T-031 门禁允许删除不属于任何模块的生成文件（DEBT-009）
- Status: Done
- Covers: AC-012-2
- Design: FLOW-02
- Flow: FLOW-02
- Scope: internal/gate/**, internal/workspace/generated.go, internal/app/gate_integration_test.go
- Depends: T-030

代理可按 cfg.02 的修复办法删除生成文件，再由单独一次 `harness init` 写回；写入其他内容仍按范围判定。

### T-032 报告失败的测试与剩余的解析诊断（DEBT-008、DEBT-002、DEBT-003）
- Status: Done
- Covers: AC-039-2, AC-003-1
- Design: MOD-05
- Scope: internal/rules/builtin/**
- Depends: T-031

新增规则：测试报告中有失败的测试时报告测试名；报告 PRD 与进度文件中尚无规则输出的解析诊断。task.08 改用 `engine.FullCheck`。

### T-033 harness init 忽略门禁运行时目录（DEBT-005）
- Status: Done
- Covers: AC-001-1
- Design: FLOW-07
- Scope: internal/scaffold/**, internal/app/init_integration_test.go
- Depends: T-032

init 写入 `.harness/.gitignore`，忽略门禁的 `snap/`、`session/` 与它自身：不改项目根目录的 `.gitignore`，不需要提交；已有项目重新运行 `harness init` 即补上。

### T-034 测试包构建失败记为失败结果（DEBT-011）
- Status: Done
- Covers: AC-020-2
- Design: IF-008
- Scope: internal/testreport/**
- Depends: T-033

`go test -json` 中某个包失败、而该包没有任何失败的测试时（构建失败、TestMain 失败、测试外的 panic、超时），解析结果增加一条该包的失败记录，由 test.03 报告。

### T-035 门禁解析命令时跟随 cd（DEBT-012）
- Status: Done
- Covers: AC-012-2
- Design: FLOW-02
- Flow: FLOW-02
- Scope: internal/gate/**, internal/app/gate_integration_test.go
- Depends: T-034

同一条命令中字面量的 `cd <目录>` 之后，相对路径按该目录解析；`cd` 的目标无法静态判定时，之后的写入交给命令后比对。

### T-036 secret.entropy 不把由单词拼成的标识符判为密钥（DEBT-013）
- Status: Done
- Covers: AC-025-2
- Design: FLOW-02
- Scope: internal/gate/**, internal/app/security_integration_test.go
- Depends: T-035

计算熵之前排除标识符形状的 token：只含字母、数字、下划线，按大小写边界、字母数字边界与 `_` 切开后至少 3 段，且大部分字母落在 3 个字符以上的段中。随机密钥切开后多为一两个字符的碎片，仍按熵判定。

### T-037 需求提交使用 Change 尾注
- Status: Done
- Covers: AC-013-5
- Design: FLOW-09, IF-014
- Flow: FLOW-01
- Scope: internal/rules/builtin/commit_01.go, internal/rules/builtin/testdata/commit.01/**, internal/app/commit_integration_test.go
- Depends: T-036

commit.01 接受只改 `docs/` 与 `harness.lock` 的提交用一个 `Change: CHG-n` 尾注代替 Task 尾注；Change 尾注引用的 CHG 须在变更日志中存在，不能与 Task 尾注同时出现。

### T-038 需求取代关系（Replaces / Replaced-By）
- Status: Done
- Covers: AC-003-6
- Design: FLOW-01, IF-014, IF-022
- Flow: FLOW-01
- Scope: internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/rules/builtin/prd_02.go, internal/rules/builtin/prd_12.go, internal/rules/builtin/trace_01.go, internal/rules/builtin/task_02.go, internal/rules/builtin/test_02.go, internal/rules/builtin/testdata/prd.12/**, internal/rules/builtin/trace_03.go, internal/rules/builtin/testdata/trace.03/**, internal/app/prd_integration_test.go, internal/app/testdata/prd/**
- Depends: T-037

被取代的条目在取代者的 Release 不大于 Current-Release 时按 Removed 判定：覆盖检查不再要求它，trace.03 报告对它的引用。新增 prd.12 校验 `Replaces` 与 `Replaced-By` 互相对应、引用存在、被取代者的 Release 小于取代者。


### T-039 门禁以项目根判定，命令从代理的当前目录解析（DEBT-015）
- Status: Done
- Covers: AC-012-2, AC-025-4
- Design: FLOW-02, FLOW-03, IF-017, IF-018, IF-019
- Flow: FLOW-02
- Scope: internal/cli/gate.go, internal/app/gate.go, internal/gate/**, internal/adapter/**, internal/app/gate_integration_test.go
- Depends: T-037

适配器在 Request 中填写 Root（从 hook 输入的 cwd 向上查找的项目根）与 Cwd；`harness gate` 以 Root 加载快照、保存命令前快照与会话状态，Root 为空时退回进程当前目录；命令中的相对路径从 Cwd 开始解析。回归测试：shell 停在子目录时，写 `../harness.lock` 被拒绝，命令前快照建在仓库根的 `.harness/` 下。

## Release 2

各任务实现新接口时，在实现处加 `// harness:IF-n` 锚点（formats.md 第 8 节），并为接口写测试名含其编号的单元测试（如 `TestUnits_IF_027`）；任务涉及已有模块时，顺带补齐这些模块中接口的锚点与单元测试。T-048、T-049 的规则依赖这些锚点与测试。

### T-040 版本信息
- Status: Done
- Covers: FR-049
- Design: MOD-18, IF-037, FLOW-07
- Flow: FLOW-07
- Scope: internal/version/**, internal/cli/**, .goreleaser.yaml, internal/scaffold/scaffold.go, internal/scaffold/rule.go, internal/scaffold/*_test.go, internal/app/version_integration_test.go
- Depends: T-039

`harness version` 与 `--version` 输出版本；发布构建经 -ldflags 注入版本号，`go install` 构建取模块版本或提交哈希（含 -dirty）。顺带为 IF-020、IF-021 补锚点与单元测试。

### T-041 按单位批准
- Status: Done
- Covers: AC-047-1, AC-047-2, AC-047-3
- Design: MOD-15, IF-027, IF-028, IF-029, IF-006, IF-011, IF-024, FLOW-08
- Flow: FLOW-08
- Scope: internal/approval/**, internal/lock/**, internal/rules/engine/**, internal/rules/builtin/approve_01.go, internal/rules/builtin/task_07.go, internal/rules/builtin/testdata/approve.01/**, internal/rules/builtin/testdata/task.07/**, internal/rules/builtin/registry.go, internal/app/**, internal/cli/**
- Depends: T-040

批准单位与哈希、批准链与状态（IF-027、IF-028），锁文件第 2 版的 `units`，`harness approve <路径、prd 或单位编号>` 显示差异并按单位记录。engine 的告警（Violation.Warn，不影响退出码）在此加入；approve.01 在锁文件为第 2 版时按单位列出未批准与失效项（告警），task.07 在第 2 版时按批准链判定。

### T-042 锁文件迁移
- Status: Done
- Covers: AC-047-6
- Design: IF-029, IF-006, IF-023, FLOW-10
- Flow: FLOW-10
- Scope: internal/approval/**, internal/lock/**, internal/app/**
- Depends: T-041

`harness lock` 把 Release 1 的按文件批准迁移为按单位批准，写审计记录；`phase` 字段由 T-054 删除（在此之前门禁仍按阶段判定产物写入）。在本仓库执行迁移后，任务需要按单位批准才能开工。

### T-043 门禁按批准链判定
- Status: Done
- Covers: AC-047-4, AC-047-5, AC-048-2, AC-048-3
- Design: IF-017, IF-028, IF-025, FLOW-02, FLOW-06
- Flow: FLOW-02
- Scope: internal/gate/**, internal/adapter/**, internal/approval/**, internal/phase/**, internal/app/**
- Depends: T-042

锁文件为第 2 版时：`docs/` 任何时候可写（去掉 gate.artifact），源码写入检查 InProgress 任务的批准链，会话开始注入批准状态摘要与建议的 skill。顺带为 IF-015 ~ IF-019、IF-025 补锚点与单元测试。

### T-044 声明式自定义规则
- Status: Done
- Covers: FR-054
- Design: MOD-17, IF-036, IF-013, IF-024, FLOW-08
- Flow: FLOW-08
- Scope: internal/rules/custom/**, internal/workspace/**, internal/lock/**, internal/app/**
- Depends: T-043

按 `harness.json` 的 rules 构造 required、enum、ref、forbid 四类规则；校验编号前缀、样本与批准单位类型；样本目录纳入锁文件；`approveFor` 的规则在批准该类单位时运行。

### T-045 外部工具运行器与 lint
- Status: Done
- Covers: AC-014-2, AC-014-3, AC-014-4, AC-014-5
- Design: MOD-16, IF-030, IF-031, IF-032, IF-033, IF-034, IF-038, FLOW-11
- Flow: FLOW-11
- Scope: internal/toolrun/**, internal/lang/**, internal/testreport/**, internal/workspace/**, internal/app/**, internal/cli/**, internal/rules/builtin/lint_01.go, internal/rules/builtin/cfg_03.go, internal/rules/builtin/testdata/lint.01/**, internal/rules/builtin/testdata/cfg.03/**, internal/rules/builtin/registry.go
- Depends: T-044

五种语言的语言配置包（IF-034）；toolrun 的 Lint、Unit、Smoke 运行器与报告读写；`harness check lint|unit|smoke`；规则 lint.01、cfg.03。Unit 的覆盖率采集在 T-049 接入 Go。

### T-046 启动与冒烟测试命令声明
- Status: Done
- Covers: AC-055-1
- Design: IF-010, IF-014, FLOW-01
- Flow: FLOW-01
- Scope: internal/workspace/**, internal/rules/builtin/smoke_01.go, internal/rules/builtin/testdata/smoke.01/**, internal/rules/builtin/registry.go, internal/app/**
- Depends: T-045

项目配置的 `start` 字段（`none: <原因>`）与规则 smoke.01；本仓库的 `harness.json` 补上 start（由人修改并登记 CHG）。

### T-047 init 与 CI：语言配置包、工具配置文件、固定版本
- Status: Done
- Covers: AC-014-1, AC-014-6, AC-014-7, AC-055-2, FR-051
- Design: IF-020, IF-021, IF-026, IF-034, IF-037, FLOW-07
- Flow: FLOW-07
- Scope: internal/scaffold/**, internal/lang/**, internal/app/**, internal/cli/**
- Depends: T-046

init 改用 IF-034 的语言配置包，缺失时生成工具配置文件；CI 模板依次运行 lint、unit、smoke、check、check commits；用户项目下载固定版本的发布归档并校验，没有发布版本号时拒绝（`--ci-version`）；cfg.02 按 CI 固定版本比对并提示升级。本仓库的 CI 配置用 `harness init` 重新生成。

### T-048 Go 接口锚点与签名快照
- Status: InProgress
- Covers: FR-052
- Design: IF-035, IF-039, IF-014, FLOW-01, FLOW-12
- Flow: FLOW-01
- Scope: internal/lang/**, internal/app/**, internal/cli/**, internal/workspace/**, internal/lock/**, internal/model/**, internal/rules/engine/**, internal/rules/builtin/**, internal/rules/custom/**, internal/vcs/**, internal/testreport/**, internal/approval/**
- Depends: T-047

Go Analyzer 的 Imports、Anchors、Resolve；`harness snapshot` 写 `docs/architecture/interfaces.txt`；规则 if.01 ~ if.04；锁文件的 `snapshot:IF-n`。补齐 FLOW-01 各模块的锚点。详细设计与代码签名不一致的接口，改设计或改代码，改设计时需人重新批准。

### T-049 单元测试追溯
- Status: Todo
- Covers: FR-021
- Design: IF-031, IF-033, IF-035, IF-014, FLOW-11
- Flow: FLOW-11
- Scope: internal/lang/**, internal/toolrun/**, internal/testreport/**, internal/workspace/**, internal/app/**, internal/cli/**, internal/rules/builtin/unit_0*.go, internal/rules/builtin/testdata/unit.0*/**, internal/rules/builtin/registry.go
- Depends: T-048

Go Analyzer 的 Coverage；Unit 运行器按接口运行测试并统计模块覆盖率；规则 unit.01 ~ unit.03。补齐 IF-004、IF-022 ~ IF-026 的单元测试；模块覆盖率低于阈值时补测试。

### T-050 Go 模块依赖与封装
- Status: Todo
- Covers: FR-053, FR-035
- Design: IF-035, IF-014, FLOW-01
- Flow: FLOW-01
- Scope: internal/lang/**, internal/rules/builtin/deps_0*.go, internal/rules/builtin/testdata/deps.0*/**, internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/app/**
- Depends: T-049

规则 deps.01 ~ deps.04：跨模块 import 须在 Depends 中且指向 Public，sourcePaths 内的源文件须有归属，Depends 中的模块须被使用。

### T-051 机械改动任务
- Status: Todo
- Covers: FR-050
- Design: IF-035, IF-014, FLOW-09
- Flow: FLOW-09
- Scope: internal/lang/**, internal/rules/builtin/task_03.go, internal/rules/builtin/task_05.go, internal/rules/builtin/task_09.go, internal/rules/builtin/commit_03.go, internal/rules/builtin/commit_05.go, internal/rules/builtin/testdata/task.09/**, internal/rules/builtin/testdata/commit.05/**, internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/app/**
- Depends: T-050

`Kind: mechanical` 与 Rewrite 的声明（task.09）；task.03、task.05、commit.03 豁免机械改动任务；commit.05 逐行校验只改 import 路径与模块声明行。

### T-052 文档引用检查
- Status: Todo
- Covers: FR-030
- Design: IF-012, IF-014, IF-022, IF-009, FLOW-01
- Flow: FLOW-01
- Scope: internal/rules/builtin/doc_0*.go, internal/rules/builtin/testdata/doc.0*/**, internal/rules/builtin/registry.go, internal/rules/builtin/util.go, internal/workspace/**, internal/app/**, internal/cli/**
- Depends: T-051

规则 doc.02 ~ doc.05：受检 Markdown 行内代码中的路径、`harness` 子命令、编号引用，以及详细设计可能过期的告警；cli 把子命令名传入 CheckOptions。修正本仓库文档中被报告的过期引用。

### T-053 抑制注释登记
- Status: Todo
- Covers: FR-042
- Design: IF-003, IF-014, FLOW-01
- Flow: FLOW-01
- Scope: internal/rules/builtin/supp_01.go, internal/rules/builtin/testdata/supp.01/**, internal/rules/builtin/registry.go, internal/app/**
- Depends: T-052

规则 supp.01：抑制注释同一行须引用 Open 的 DEBT。

### T-054 删除阶段（门禁、命令）
- Status: Todo
- Covers: AC-048-1
- Design: IF-017, IF-024, FLOW-02
- Flow: FLOW-02
- Scope: internal/gate/**, internal/app/**, internal/cli/**, internal/phase/**
- Depends: T-053

去掉门禁、app、cli 中的阶段逻辑与 Release 1 的按文件批准分支；`harness phase` 提示已删除。

### T-055 删除阶段模块，Release 2 生效
- Status: Todo
- Covers: FR-048
- Design: MOD-10, EXT-04, IF-020, FLOW-07
- Flow: FLOW-07
- Scope: internal/phase/**, internal/scaffold/**, internal/app/**, internal/cli/**
- Depends: T-054

删除 `internal/phase/`、scaffold 中的阶段用法与 skill 中的阶段说明；设计中 MOD-10、EXT-04 改为 Removed，去掉对 MOD-10 的依赖与 IF-015、IF-016 的流程步骤；PRD 的 Current-Release 改为 2。设计与 PRD 的这些改动需人批准。

### T-056 默认工具命令按语言配置包的输出格式解析
- Status: Todo
- Covers: AC-014-3, AC-014-4
- Design: IF-030, IF-034
- Scope: internal/toolrun/**
- Depends: T-049

`harness init` 把各语言的默认命令写入 `tools`（AC-014-1），toolrun 却把 `tools` 中的命令一律按 `lines` 解析，`gofmt -l` 这类按文件名输出的默认命令因此查不出问题。`tools` 中的命令与语言配置包的默认命令相同时，沿用配置包声明的输出格式与安装提示。
