# 架构总览：harness

- Status: Approved
- Updated: 2026-11-26
- PRD: 1.0-rc3

本文件是概要设计：模块划分、依赖、扩展点。每个模块的接口签名和内部设计见 `docs/architecture/modules/`，跨模块调用流程见 `docs/architecture/flows.md`，产物格式见 `docs/architecture/formats.md`，架构决策见 `docs/adr/`。

## 1. 分层与依赖

```
L4 入口   MOD-01 cli ──▶ app, adapter, version
L3 编排   MOD-02 app
L2 能力   MOD-13 scaffold ──▶ adapter, lang, version
          MOD-11 gate ──────▶ approval, engine            MOD-12 adapter ──▶ gate ◀ EXT-02
          MOD-16 toolrun ───▶ testreport, lang
          MOD-17 custom ────▶ engine                      （项目自定义规则）
          MOD-05 builtin ───▶ engine, approval + L0        ◀ EXT-03
          MOD-04 engine
          MOD-15 approval ──▶ workspace, lock, model
L1 加载   MOD-09 workspace
L0 基础   MOD-03 model · MOD-06 lock · MOD-07 vcs · MOD-08 testreport · MOD-14 lang ◀ EXT-01 · MOD-18 version
```

依赖规则（与第 2 节各模块的 Depends 字段一致，以 Depends 字段为准）：

- 叶子模块（model、lock、vcs、testreport、lang、version）不依赖任何模块。
- workspace 只依赖叶子模块，负责一次性加载项目快照。
- engine 只依赖 workspace；approval 依赖 workspace 与叶子模块；builtin（内置规则）依赖 engine、approval 和叶子模块；custom 依赖 engine。
- gate、scaffold、toolrun 依赖其下各层，不互相依赖。
- app 是唯一的编排层；cli 只依赖 app、adapter 和 version。

## 2. 模块

### MOD-01 cli
- Status: Active
- Paths: cmd/harness/**, internal/cli/**, go.mod, go.sum, .goreleaser.yaml, .github/workflows/release.yml
- Public: cmd/harness/main.go
- Depends: MOD-02, MOD-04, MOD-11, MOD-12, MOD-13, MOD-18
- Interfaces:
- Covers: FR-001, FR-007, FR-012, FR-037, FR-048, FR-049, NFR-002
- Core: false
- Coverage: 60

命令行入口：参数解析、输出格式（文本 / `--json`）、退出码（0 通过 / 1 违规 / 2 内部错误）。同时拥有 Go 模块文件和发布配置（ADR-008）。

### MOD-02 app
- Status: Active
- Paths: internal/app/**
- Public: internal/app/app.go
- Depends: MOD-04, MOD-05, MOD-06, MOD-09, MOD-11, MOD-13, MOD-15, MOD-16, MOD-17
- Interfaces: IF-022, IF-023, IF-024, IF-025, IF-026, IF-038, IF-039
- Covers: FR-001, FR-004, FR-007, FR-012, FR-014, FR-017, FR-021, FR-037, FR-047, FR-052, FR-054, FR-055
- Core: true
- Coverage: 80

用例编排：check、trace、lock、approve、门禁、init、外部工具报告（lint、unit、smoke）、接口快照。交互式终端判定在此层。代理协议转换在 cli 调用 adapter 完成，app 只接收门禁请求。

### MOD-03 model
- Status: Active
- Paths: internal/model/**
- Public: internal/model/model.go
- Depends:
- Interfaces: IF-001, IF-002
- Covers: FR-003, FR-006, FR-010, FR-011, FR-019, FR-031, FR-033
- Core: true
- Coverage: 90

Markdown 字段格式的通用解析器，以及 PRD、架构、ADR、任务、技术债、进度的统一实体模型。

### MOD-04 engine
- Status: Active
- Paths: internal/rules/engine/**
- Public: internal/rules/engine/engine.go
- Depends: MOD-09
- Interfaces: IF-011, IF-012, IF-013
- Covers: FR-028, FR-030, FR-054
- Core: true
- Coverage: 90

规则引擎：规则接口、注册表、按选择器运行、违规结构、样本自检。

### MOD-05 builtin
- Status: Active
- Paths: internal/rules/builtin/**
- Public: internal/rules/builtin/registry.go
- Depends: MOD-03, MOD-04, MOD-06, MOD-07, MOD-08, MOD-09, MOD-14, MOD-15
- Interfaces: IF-014
- Covers: FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, FR-010, FR-011, FR-013, FR-014, FR-019, FR-020, FR-021, FR-024, FR-030, FR-031, FR-033, FR-034, FR-035, FR-037, FR-039, FR-040, FR-041, FR-042, FR-043, FR-047, FR-050, FR-052, FR-053, FR-055
- Core: true
- Coverage: 90

全部内置规则，每条规则一个文件，样本在 `testdata/<规则编号>/`。

### MOD-06 lock
- Status: Active
- Paths: internal/lock/**
- Public: internal/lock/lock.go
- Depends:
- Interfaces: IF-005, IF-006
- Covers: FR-004, FR-037, FR-041, FR-047, FR-048
- Core: true
- Coverage: 90

内容规范化与哈希、`harness.lock` 读写、批准记录、当前阶段。

### MOD-07 vcs
- Status: Active
- Paths: internal/vcs/**
- Public: internal/vcs/vcs.go
- Depends:
- Interfaces: IF-007
- Covers: FR-010, FR-013, FR-034
- Core: false
- Coverage: 75

git 访问：提交范围、提交改动、指定版本的文件内容。通过调用 `git` 命令实现。

### MOD-08 testreport
- Status: Active
- Paths: internal/testreport/**
- Public: internal/testreport/testreport.go
- Depends:
- Interfaces: IF-008, IF-033
- Covers: FR-014, FR-020, FR-021, FR-039
- Core: false
- Coverage: 85

解析 JUnit XML 与 `go test -json`，输出统一的测试结果。

### MOD-09 workspace
- Status: Active
- Paths: internal/workspace/**
- Public: internal/workspace/workspace.go
- Depends: MOD-03, MOD-06, MOD-07, MOD-08, MOD-14
- Interfaces: IF-009, IF-010
- Covers: FR-003, FR-014, FR-030, FR-041, FR-055
- Core: false
- Coverage: 80

加载项目快照：配置（`harness.json`）、产物模型、锁文件、git 仓库、测试报告、语言注册表。

### MOD-10 phase
- Status: Removed
- Paths: internal/phase/**
- Public: internal/phase/phase.go
- Depends: MOD-04, MOD-09
- Interfaces: IF-015, IF-016
- Covers: FR-038, FR-045
- Core: true
- Coverage: 85

阶段定义、当前阶段与推进。Release 2 取消阶段（FR-048），本模块与其代码目录一并删除。

### MOD-11 gate
- Status: Active
- Paths: internal/gate/**
- Public: internal/gate/gate.go
- Depends: MOD-03, MOD-04, MOD-09, MOD-15
- Interfaces: IF-017, IF-018
- Covers: FR-012, FR-024, FR-025, FR-037, FR-047, FR-048
- Core: true
- Coverage: 90

与代理无关的门禁判定：读写、命令、会话开始、写后检查、结束检查；Bash 命令解析；写前快照与写后回滚。

### MOD-12 adapter
- Status: Active
- Paths: internal/adapter/**
- Public: internal/adapter/adapter.go
- Depends: MOD-11
- Interfaces: IF-019
- Covers: FR-012
- Core: false
- Coverage: 85

代理适配器注册表；把各代理的 hook 输入输出转换为门禁请求和判定。Release 1 只有 `claudecode`。

### MOD-13 scaffold
- Status: Active
- Paths: internal/scaffold/**
- Public: internal/scaffold/scaffold.go
- Depends: MOD-04, MOD-09, MOD-12, MOD-14, MOD-18
- Interfaces: IF-020, IF-021
- Covers: FR-001, FR-002, FR-013, FR-014, FR-016, FR-041, FR-043, FR-048, FR-051, FR-055
- Core: false
- Coverage: 80

`go:embed` 模板（产物骨架、skill、rule、CI、git hook）；生成与"期望内容"计算（供生成文件一致性检查）。

### MOD-14 lang
- Status: Active
- Paths: internal/lang/**
- Public: internal/lang/lang.go
- Depends:
- Interfaces: IF-003, IF-004, IF-034, IF-035
- Covers: FR-014, FR-020, FR-021, FR-031, FR-050, FR-052, FR-053
- Core: false
- Coverage: 85

语言注册表：文件扩展名、注释提取、测试分类（单元 / 集成）、语言配置包（格式化与静态检查命令）；Go 另有源码分析（import、接口锚点与签名、覆盖率）。

### MOD-15 approval
- Status: Active
- Paths: internal/approval/**
- Public: internal/approval/approval.go
- Depends: MOD-03, MOD-06, MOD-09
- Interfaces: IF-027, IF-028, IF-029
- Covers: FR-047, FR-048
- Core: true
- Coverage: 85

批准单位、内容哈希、批准链与批准状态、批准记录与 Release 1 锁文件迁移（ADR-009）。

### MOD-16 toolrun
- Status: Active
- Paths: internal/toolrun/**
- Public: internal/toolrun/toolrun.go
- Depends: MOD-08, MOD-09, MOD-14
- Interfaces: IF-030, IF-031, IF-032
- Covers: FR-014, FR-021, FR-055
- Core: false
- Coverage: 75

运行外部工具（格式化与静态检查、按接口的单元测试与覆盖率、冒烟测试）并写报告；全量检查只读报告（ADR-011）。

### MOD-17 custom
- Status: Active
- Paths: internal/rules/custom/**
- Public: internal/rules/custom/custom.go
- Depends: MOD-04, MOD-09
- Interfaces: IF-036
- Covers: FR-054
- Core: false
- Coverage: 85

项目声明式自定义规则：按 `harness.json` 的 rules 构造规则，校验编号、样本与批准前检查登记（ADR-010）。

### MOD-18 version
- Status: Active
- Paths: internal/version/**
- Public: internal/version/version.go
- Depends:
- Interfaces: IF-037
- Covers: FR-049, FR-051
- Core: false
- Coverage: 80

当前二进制的版本与发布版本判定。

## 3. 扩展点

### EXT-01 新增编程语言
- Status: Active
- Interface: IF-003
- Paths: internal/lang/<语言>/**, internal/lang/registry.go, internal/scaffold/templates/langs/<语言>/**
- Closed: MOD-03, MOD-05, MOD-11
- Relates: NFR-006

### EXT-02 新增编程代理
- Status: Active
- Interface: IF-019
- Paths: internal/adapter/<代理>/**, internal/scaffold/templates/agents/<代理>/**, internal/adapter/registry.go
- Closed: MOD-11
- Relates: NFR-007

### EXT-03 新增内置规则
- Status: Active
- Interface: IF-011
- Paths: internal/rules/builtin/<规则文件>.go, internal/rules/builtin/testdata/<规则编号>/**, internal/rules/builtin/registry.go
- Closed: MOD-03, MOD-04, MOD-11
- Relates: NFR-009

### EXT-04 新增内置阶段
- Status: Removed
- Interface: IF-015
- Paths: internal/phase/builtin/<阶段>.json, internal/phase/builtin/order.json, internal/scaffold/templates/skills/<阶段>/**
- Closed: MOD-03, MOD-04, MOD-11
- Relates: NFR-008

Release 2 随阶段一起删除（FR-048）。

项目级扩展（不重新编译）：自定义规则写在 `harness.json` 的 `rules`，由 MOD-17 构造为 IF-011 规则（FR-054，ADR-010），可登记为某类批准单位的批准前检查。Release 1 的项目阶段配置已删除。
