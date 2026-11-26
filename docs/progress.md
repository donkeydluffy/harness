# 进度记录

- Status: Draft
- Updated: 2026-11-26

任务变为 Done 时追加一条记录（AC-024-2）：标题为 `### <日期> T-n`，正文写结论与遗留问题。会话结束检查未通过的项也记在这里。

## 记录

### 2026-11-26 T-001
- 结论：Done。`harness check [规则或前缀] [--json]` 可用，退出码 0/1/2；model、workspace、engine、prd.01、prd.02 完成。17 个用例（AC-003-1、AC-003-2、AC-003-4、NFR-004）的集成测试全部通过（`go test ./...`）；故意删掉 prd.02 的 Priority 检查后 3 个测试失败，恢复后通过。本仓库 `harness check` 0 条违规。
- 调整：AC-003-5 的追溯部分依赖 trace.01，移到 T-006 验收；T-001 范围补充 `internal/rules/builtin/util.go` 与 `registry_test.go`。
- 遗留：DEBT-001 模块路径；DEBT-002 解析诊断未输出。task.08 尚未实现，本任务的完成定义由 `go test` 人工核对，T-015 上线后回溯校验。
