---
name: plan
description: 规划阶段：把已批准的需求与设计拆成任务，写入 docs/tasks.md。
check: harness check task
---

# 任务规划

前提：PRD 与架构产物已由人批准。

## 需求或设计不清时

需求或设计不清时登记待澄清项并停止规划：在 PRD 中新增 `Status: Open` 的 OQ，用 `Relates` 写明关联的需求，告诉人需要回到需求或架构阶段澄清。不要为不清楚的需求生成任务，也不要自行假设答案。

## 任务格式

```markdown
### T-003 离线同步队列
- Status: Todo
- Covers: AC-003-1
- Design: MOD-02, IF-004
- Scope: internal/sync/**
- Depends: T-002
```

- Covers 写 FR、NFR 或 AC 编号；当前 Release 的每条需求都要被任务覆盖。
- Scope 默认落在一个模块的路径内；跨模块的任务加 `- Flow: FLOW-n`，涉及的模块都要在该流程中。
- 扩展已有扩展点的任务加 `- Extends: EXT-n`，Scope 不能包含该扩展点的封闭模块。
- Depends 无环，被依赖的任务排在前面。
- 一个任务的改动应能拆成每次不超过 400 行的提交。

## 结束

运行 `harness check task`，修复全部违规。告诉人可以评审任务列表。
