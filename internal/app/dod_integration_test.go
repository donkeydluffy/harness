package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// approveFiles 在伪终端中由"人"依次批准产物。
func approveFiles(t *testing.T, dir string, artifacts ...string) {
	t.Helper()
	for _, a := range artifacts {
		if out, code := inTerminal(t, dir, confirmPrompt, "y", "approve", a); code != 0 {
			t.Fatalf("批准 %s 失败：%d %s", a, code, out)
		}
	}
}

func write(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

var inProgress = taskDoc(task("T-001", "InProgress", "FR-001", "MOD-01", "internal/api/**", "", ""),
	task("T-002", "Todo", "FR-002", "MOD-02", "internal/store/**", "T-001", ""))

// gateRepo 返回 PRD 与架构总览均已批准、T-001 为 InProgress 的仓库。
func gateRepo(t *testing.T) string {
	t.Helper()
	dir := lockedRepo(t, basePRD)
	write(t, dir, "docs/tasks.md", inProgress)
	approveFiles(t, dir, "prd", "docs/architecture/overview.md")
	return dir
}

// ---- AC-038-1 PRD 有效批准 ----

func TestCheck_AC_038_1_N1(t *testing.T) {
	wantPass(t, checkJSON(t, gateRepo(t), "task.07"))
}

func TestCheck_AC_038_1_E1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	write(t, dir, "docs/tasks.md", inProgress)
	approveFiles(t, dir, "docs/architecture/overview.md")
	wantViolation(t, checkJSON(t, dir, "task.07"), "task.07", "T-001", "PRD 没有处于有效的批准状态")
}

func TestCheck_AC_038_1_E2(t *testing.T) {
	dir := gateRepo(t)
	edit(t, dir, "- AC-001-1: 生成骨架。", "- AC-001-1: 生成骨架与配置。")
	wantViolation(t, checkJSON(t, dir, "task.07"), "task.07", "T-001", "PRD 没有处于有效的批准状态")
}

// ---- AC-038-2 设计产物有效批准 ----

func TestCheck_AC_038_2_N1(t *testing.T) {
	wantPass(t, checkJSON(t, gateRepo(t), "task.07"))
}

func TestCheck_AC_038_2_E1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	write(t, dir, "docs/tasks.md", inProgress)
	approveFiles(t, dir, "prd")
	wantViolation(t, checkJSON(t, dir, "task.07"), "task.07", "MOD-01", "docs/architecture/overview.md")
}

// ---- AC-038-3 / AC-011-5 关联需求没有 Open 的待澄清项 ----

const oq003 = "### OQ-003 接口格式\n- Status: Open\n- Relates: FR-001\n\n## 8. 变更日志"

func TestCheck_AC_038_3_N1(t *testing.T) {
	wantPass(t, checkJSON(t, gateRepo(t), "task.07")) // OQ-001 已 Closed
}

func TestCheck_AC_038_3_E1(t *testing.T) {
	dir := gateRepo(t)
	edit(t, dir, "## 8. 变更日志", oq003)
	wantViolation(t, checkJSON(t, dir, "task.07"), "task.07", "T-001", "FR-001", "OQ-003")
}

func TestCheck_AC_011_5_E1(t *testing.T) {
	dir := gateRepo(t)
	edit(t, dir, "## 8. 变更日志", oq003)
	wantViolation(t, checkJSON(t, dir, "tasks"), "task.07", "OQ-003")
}

// ---- AC-039-1、AC-039-2、AC-024-2 完成定义 ----

const record = "# 进度\n\n### 2026-11-26 T-001\n- 结论：Done，全部用例通过。\n- 遗留：无。\n"

// doneRepo 返回 T-001 为 Done 的仓库：covers 为 T-001 的关联条目，tests 为报告中 passed 的测试。
func doneRepo(t *testing.T, prdText, covers string, tests ...string) string {
	t.Helper()
	dir := lockedRepo(t, prdText)
	write(t, dir, "docs/tasks.md", taskDoc(task("T-001", "Done", covers, "MOD-01", "internal/api/**", "", ""),
		task("T-002", "Todo", "FR-001, FR-002", "MOD-02", "internal/store/**", "T-001", "")))
	write(t, dir, "docs/progress.md", record)
	write(t, dir, "internal/api/api_integration_test.go", goTests(tests...))
	write(t, dir, ".harness/reports/go.jsonl", goReport(allPass(tests...)...))
	return dir
}

var fr001Tests = []string{"TestX_AC_001_1_N1", "TestX_AC_001_1_E1"}

func TestCheck_AC_039_1_N1(t *testing.T) {
	wantPass(t, checkJSON(t, doneRepo(t, basePRD, "FR-001", fr001Tests...), "task.08"))
}

func TestCheck_AC_039_1_E1(t *testing.T) {
	dir := doneRepo(t, basePRD, "FR-001, FR-002", append(fr001Tests, "TestX_AC_002_1_N1")...)
	v := wantViolation(t, checkJSON(t, dir, "task.08"), "task.08", "T-001", "AC-002-1/B1")
	if strings.Contains(v.Msg, "AC-001-1") {
		t.Fatalf("已通过的用例不应列出：%s", v.Msg)
	}
}

func TestCheck_AC_039_1_B1(t *testing.T) {
	src := prd(t, "  - AC-001-1/E1: 只读目录", "  - AC-001-1/E1: Removed：不再考虑只读目录\n  - AC-001-1/B1: 路径含空格")
	wantPass(t, checkJSON(t, doneRepo(t, src, "FR-001", "TestX_AC_001_1_N1", "TestX_AC_001_1_B1"), "task.08"))
}

func TestCheck_AC_039_1_B2(t *testing.T) {
	// 只关联 AC-001-1；同一 FR 下 AC-001-2 的用例没有测试也不影响 T-001。
	wantPass(t, checkJSON(t, doneRepo(t, twoACs(t), "AC-001-1", fr001Tests...), "task.08"))
}

func TestCheck_AC_039_2_N1(t *testing.T) {
	wantPass(t, checkJSON(t, doneRepo(t, basePRD, "FR-001", fr001Tests...), "task.08"))
}

func TestCheck_AC_039_2_E1(t *testing.T) {
	dir := doneRepo(t, basePRD, "FR-001", fr001Tests...)
	write(t, dir, "docs/debt.md", "# 技术债清单\n\n### DEBT-001 x\n- Status: Done\n- Module: MOD-01\n\n说明。\n")
	wantViolation(t, checkJSON(t, dir, "task.08"), "task.08", "全量检查", "T-001")
}

func TestCheck_AC_024_2_N1(t *testing.T) {
	wantPass(t, checkJSON(t, doneRepo(t, basePRD, "FR-001", fr001Tests...), "task.08"))
}

func TestCheck_AC_024_2_E1(t *testing.T) {
	dir := doneRepo(t, basePRD, "FR-001", fr001Tests...)
	write(t, dir, "docs/progress.md", "# 进度\n\n### 2026-11-26 T-002\n- 结论：进行中。\n- 遗留：无。\n")
	wantViolation(t, checkJSON(t, dir, "task.08"), "task.08", "T-001", "没有该任务的记录")
}

func TestCheck_AC_024_2_E2(t *testing.T) {
	dir := doneRepo(t, basePRD, "FR-001", fr001Tests...)
	write(t, dir, "docs/progress.md", strings.Replace(record, "- 结论：Done，全部用例通过。\n", "", 1))
	v := wantViolation(t, checkJSON(t, dir, "task.08"), "task.08", "T-001", "缺少：结论")
	if v.File != "docs/progress.md" || v.Line != 3 {
		t.Fatalf("位置：%s:%d", v.File, v.Line)
	}
}
