package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const t002Todo = "# 任务列表\n\n- Status: Draft\n\n### T-002 存储\n- Status: Todo\n- Covers: FR-002\n- Design: MOD-02\n- Scope: internal/store/**\n- Depends:\n"

// chainRepo 返回锁文件为第 2 版、PRD、MOD-02 与 Todo 任务 T-002 已按单位批准的仓库。
func chainRepo(t *testing.T) string {
	t.Helper()
	dir := unitRepo(t)
	writeFile(t, dir, "docs/tasks.md", t002Todo)
	writeFile(t, dir, "internal/store/store.go", "package store\n")
	mustApprove(t, dir, "prd", "MOD-02", "T-002")
	return dir
}

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func start(t *testing.T, dir, id string) {
	t.Helper()
	editFile(t, dir, "docs/tasks.md", "### "+id+" 存储\n- Status: Todo", "### "+id+" 存储\n- Status: InProgress")
}

func writeStore(t *testing.T, dir string) hookOut {
	t.Helper()
	return hook(t, dir, "PreToolUse", writeTool(dir, "internal/store/store.go", "package store\n\nvar X = 1\n"))
}

// ---- AC-047-4 任务开工与源码写入按批准链判定 ----

func TestGate_AC_047_4_N1(t *testing.T) {
	dir := chainRepo(t)
	start(t, dir, "T-002")
	wantPass(t, checkJSON(t, dir, "task.07"))
	wantAllow(t, writeStore(t, dir))
}

func TestGate_AC_047_4_E1(t *testing.T) {
	dir := unitRepo(t)
	writeFile(t, dir, "docs/tasks.md", t002Todo)
	mustApprove(t, dir, "prd", "MOD-02")
	start(t, dir, "T-002")
	wantViolation(t, checkJSON(t, dir, "task.07"), "task.07", "T-002", "未批准")
}

func TestGate_AC_047_4_E2(t *testing.T) {
	dir := chainRepo(t)
	start(t, dir, "T-002")
	editFile(t, dir, "docs/PRD.md", "- AC-002-1: 记录问答。", "- AC-002-1: 记录问答与结论。")
	wantDeny(t, writeStore(t, dir), "gate.write", "T-002", "FR-002 批准已失效")
}

func TestGate_AC_047_4_E3(t *testing.T) {
	dir := chainRepo(t)
	editFile(t, dir, "docs/PRD.md", "- Resolution: 支持\n", "- Resolution: 支持\n\n### OQ-003 问答保留多久\n- Status: Open\n- Relates: FR-002\n")
	start(t, dir, "T-002")
	wantViolation(t, checkJSON(t, dir, "task.07"), "task.07", "T-002", "OQ-003")
}

// ---- AC-047-5 代理随时可写草稿，修改已批准单位使其失效 ----

func TestGate_AC_047_5_N1(t *testing.T) {
	dir := chainRepo(t)
	start(t, dir, "T-002")
	wantAllow(t, hook(t, dir, "PreToolUse", editTool(dir, "docs/PRD.md", "## 4. 非功能需求", fr003)))
	addFR003(t, dir)
	editFile(t, dir, "docs/tasks.md", "- Depends:\n", "- Depends:\n\n### T-003 导出\n- Status: InProgress\n- Covers: FR-003\n- Design: MOD-02\n- Scope: internal/store/**\n- Depends:\n")
	wantViolation(t, checkJSON(t, dir, "task.07"), "task.07", "T-003", "FR-003 未批准")
}

func TestGate_AC_047_5_E1(t *testing.T) {
	dir := chainRepo(t)
	start(t, dir, "T-002")
	editFile(t, dir, "docs/tasks.md", "- Scope: internal/store/**", "- Scope: internal/store/**, internal/api/**")
	wantDeny(t, writeStore(t, dir), "gate.write", "T-002 批准已失效")
}

// ---- AC-048-2 docs/ 任何时候可写 ----

func TestGate_AC_048_2_N1(t *testing.T) {
	dir := unitRepo(t)
	wantAllow(t, hook(t, dir, "PreToolUse", writeTool(dir, "docs/tasks.md", t002Todo)))
	writeFile(t, dir, "docs/tasks.md", strings.Replace(t002Todo, "- Status: Todo", "- Status: InProgress", 1))
	wantViolation(t, checkJSON(t, dir, "task.07"), "task.07", "T-002", "未批准")
}

func TestGate_AC_048_2_E1(t *testing.T) {
	dir := unitRepo(t)
	wantDeny(t, hook(t, dir, "PreToolUse", editTool(dir, "docs/PRD.md", "- Status: Draft\n- Current-Release", "- Status: Approved\n- Current-Release")), "gate.protect")
}

// ---- AC-048-3 会话开始注入批准状态摘要 ----

func TestGate_AC_048_3_N1(t *testing.T) {
	dir := chainRepo(t)
	mustApprove(t, dir, "docs/architecture/overview.md", "docs/architecture/modules/modules.md", "docs/architecture/flows.md", "docs/tasks.md")
	addFR003(t, dir)
	editFile(t, dir, "docs/PRD.md", "## 4. 非功能需求", strings.Replace(strings.Replace(fr003, "FR-003 导出", "FR-004 导入", 1), "AC-003-1", "AC-004-1", -1))
	h := startHook(t, dir)
	for _, want := range []string{"2 个单位未批准，0 个批准已失效", "可开工任务：T-002", "建议使用 skill：clarify"} {
		if !strings.Contains(h.Reason, want) {
			t.Errorf("注入内容应包含 %q：\n%s", want, h.Reason)
		}
	}
	if strings.Contains(h.Reason, "阶段：") {
		t.Errorf("不应再注入阶段：\n%s", h.Reason)
	}
}

func TestGate_AC_048_3_B1(t *testing.T) {
	dir := writeTree(t, map[string]string{"harness.lock": `{"version": 2}` + "\n"})
	if h := startHook(t, dir); !strings.Contains(h.Reason, "无已批准的需求，建议 clarify") {
		t.Fatalf("空项目应建议 clarify：\n%s", h.Reason)
	}
}
