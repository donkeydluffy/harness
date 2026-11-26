# 技术债清单

- Status: Draft
- Updated: 2026-11-26

每项格式见 `docs/architecture/formats.md` 第 3 节：`### DEBT-nnn 标题`，字段 Status（Open / Resolved）、Module。

## 技术债

### DEBT-001 Go 模块路径暂为 harness
- Status: Open
- Module: MOD-01

仓库还没有远程地址，`go.mod` 的模块路径暂用 `harness`。T-027 发布前改为仓库地址，否则 `go install` 不可用。

### DEBT-002 解析诊断未被任何规则报告
- Status: Open
- Module: MOD-05

`workspace.Snapshot.Diags`（代码块未闭合、用例不在验收标准下、用例编号重复）目前没有规则输出，问题会被静默忽略。T-003 起 prd.07 报告用例相关诊断，T-005 起 arch.01 报告架构文档的诊断。adr.01（T-007）与 debt.01（T-009）已报告各自文件的诊断。任务列表的诊断仍未报告，计划由 task.01（T-011）报告。
