---
name: change
description: 需求变更：修改已有需求、新增或删除需求时使用，保证每处变化都登记在变更日志中。
check: harness check prd
---

# 需求变更

PRD 已经执行过 `harness lock` 后，每处需求变化都要在 `## 变更日志` 登记，`harness check prd` 会对比锁文件找出未登记的变化。

## 步骤

1. 改需求、验收标准或用例。
2. 在变更日志表格追加一行，一次变更的多行共用同一个 CHG 编号：

   ```markdown
   | CHG-004 | 2026-02-01 | 修改 | FR-003 | 同步时限从 60 秒改为 30 秒 |
   ```

   动作取值：新增、修改、删除。动作必须与实际差异一致：Status 改为 Removed 记"删除"，新编号记"新增"。
3. 删除不删编号：需求把 Status 改为 Removed；用例在文字前加 `Removed`。被删除的编号不能再分配给新内容。
4. 受影响的设计和任务一起更新：`harness trace --format md` 列出每条需求关联的设计元素。
5. 运行 `harness check prd`，通过后执行 `harness lock` 更新锁文件。

已批准的 PRD 改动后批准自动失效，需要人重新执行 `harness approve prd`。
