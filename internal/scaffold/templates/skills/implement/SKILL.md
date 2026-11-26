---
name: implement
description: 实现阶段：按任务列表逐个实现任务，直到任务满足完成定义。
check: harness check
---

# 编码实现

## 开始一个任务

1. 选依赖都已 Done 的第一个 Todo 任务，把 Status 改为 InProgress。门禁只允许写入 InProgress 任务 Scope 内的源码。
2. 读任务 Covers 的验收标准与用例、Design 引用的接口和流程。

## 实现

- 先按 test skill 为每个用例写集成测试，再写实现。
- 需要改 Scope 外的文件时，先改 `docs/tasks.md` 中该任务的 Scope，并说明原因。
- 每次提交不超过 400 行（不含 `docs/` 与生成文件），只涉及当前任务，提交信息带 `Task: T-n` 尾注。
- `TODO`、`FIXME` 注释在同一行引用 Open 的 DEBT-n 或未完成的 T-n；暂时不修的问题登记到 `docs/debt.md`。
- 运行测试并把报告写到 `.harness/reports/`（格式见 `docs/lang/<语言>.md`）。

## 完成定义

任务改为 Done 前：

1. 任务 Covers 涉及的用例都有通过的集成测试。
2. `harness check` 通过。
3. `docs/progress.md` 追加记录：

   ```markdown
   ### 2026-02-03 T-003
   - 结论：Done。……
   - 遗留：无。
   ```

做不到时保持 InProgress，在进度记录中写明卡在哪里。
