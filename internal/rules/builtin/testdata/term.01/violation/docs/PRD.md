# PRD: sample

- Status: Draft
- Current-Release: 1

## 2. 统一语言

| ID | 术语 | 代码名 | 定义 | 禁用别名 |
|---|---|---|---|---|
| TERM-001 | 产物 | Artifact | 文档 | 交付物 |

## 3. 功能需求

### FR-001 初始化
- Status: Active
- Release: 1
- Priority: P0

- AC-001-1: 生成全部交付物。
  - AC-001-1/N1: 空目录
  - AC-001-1/E1: 只读目录

## 4. 非功能需求

### NFR-001 确定性
- Status: Active
- Release: 1
- Type: Constraint
- Priority: P0

- AC-N001-1: 不联网。
  - AC-N001-1/N1: 断网
  - AC-N001-1/E1: 引入网络包

## 8. 变更日志

| ID | 日期 | 动作 | 对象 | 说明 |
|---|---|---|---|---|
| CHG-001 | 2026-11-25 | 新增 | FR-001 | 初稿 |
| CHG-001 | 2026-11-25 | 新增 | NFR-001 | 初稿 |
