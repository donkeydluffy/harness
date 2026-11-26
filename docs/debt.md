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

`workspace.Snapshot.Diags`（代码块未闭合、用例不在验收标准下、用例编号重复）目前没有规则输出，问题会被静默忽略。T-003 起 prd.07 报告用例相关诊断，T-005 起 arch.01 报告架构文档的诊断。adr.01（T-007）、debt.01（T-009）、task.01（T-011）已报告各自文件的诊断。仍未报告的：PRD 中与用例无关的诊断（如代码块未闭合、实体编号重复），以及 `docs/progress.md` 的诊断。

### DEBT-003 task.08 自行维护全量检查的排除列表
- Status: Open
- Module: MOD-05

`engine.FullCheck` 定义了"全量检查"不含的覆盖类规则（test.02、arch.05），会话结束检查与会话开始注入使用它；task.08 仍在 task_08.go 中重复列出这两条规则。T-018 的模块范围不含 MOD-05，没有一并修改。以后修改 task.08 的任务应改为调用 `engine.FullCheck("task.08")`，避免两处列表不一致。
