---
name: architect
description: 架构设计：依据已批准的需求设计模块、接口、调用流程、扩展点与架构决策。
check: harness check arch trace adr term
---

# 架构设计

前提：PRD 已由人批准（`harness check approve` 无违规）。

## 产物

- `docs/architecture/overview.md`：模块 `### MOD-n 名称`，字段 Status、Paths、Public、Depends、Interfaces、Covers、Core、Coverage；扩展点 `### EXT-01 名称`，字段 Status、Interface、Paths、Closed、Relates（Relates 指向 Scenario 类 NFR）。
- `docs/architecture/modules/*.md`：接口 `### IF-n 名称`，字段 Module、Covers，加一个首行为 `// IF-n` 的签名代码块。
- docs/architecture/flows.md：流程 `### FLOW-n 名称`，字段 Covers、Steps，步骤写作 `<调用方模块> -> <接口>, <调用方模块> -> <接口>`，如 `MOD-n -> IF-n`。
- `docs/adr/ADR-<n>-<标题>.md`：一文件一条决策，字段 Status（Proposed / Accepted / Superseded）、Relates，正文含 `## 背景`、`## 决策`、`## 后果`。

## 要求

- 当前 Release 的每条需求至少被一个模块、接口或流程覆盖；没有覆盖需求的元素是孤儿，删掉或补上 Covers。
- 依赖无环；流程中每一步调用的接口属于调用方自身或其 Depends 中的模块。
- 每个 Scenario 类 NFR 由一个扩展点承载；没有对应扩展场景的扩展点是过度设计。
- 有取舍的决定写成 ADR，包括被否决的方案。
- 领域概念的代码名登记在 PRD 统一语言表格中。

## 结束

运行 `harness check arch trace adr term`，修复全部违规。告诉人可以评审，批准请执行 `harness approve <产物路径>`。
