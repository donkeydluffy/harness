package app_test

import (
	"strings"
	"testing"
)

const debtArch = "# 架构总览\n\n- Status: Draft\n\n### MOD-01 core\n- Status: Active\n"

// debtList 生成技术债清单；参数为编号、状态、所属模块三元组。
func debtList(triples ...string) string {
	var b strings.Builder
	b.WriteString("# 技术债清单\n\n- Status: Draft\n\n## 技术债\n")
	for i := 0; i+2 < len(triples); i += 3 {
		b.WriteString("\n### " + triples[i] + " 某项债务\n- Status: " + triples[i+1] + "\n- Module: " + triples[i+2] + "\n\n说明文字。\n")
	}
	return b.String()
}

func taskList(status string) string {
	return "# 任务列表\n\n- Status: Draft\n\n### T-005 某任务\n- Status: " + status + "\n"
}

// debtRepo 写入架构、技术债清单、任务列表与一个 Go 源文件（第 3 行为给定注释）。
func debtRepo(t *testing.T, debts, tasks, src string) string {
	t.Helper()
	return writeTree(t, map[string]string{
		"docs/architecture/overview.md": debtArch,
		"docs/debt.md":                  debts,
		"docs/tasks.md":                 tasks,
		"internal/a.go":                 "package a\n\n" + src + "\nfunc F() {}\n",
	})
}

// ---- AC-031-1 技术债必填字段 ----

func TestCheck_AC_031_1_N1(t *testing.T) {
	wantPass(t, checkJSON(t, debtRepo(t, debtList("DEBT-001", "Open", "MOD-01"), taskList("Todo"), "// 普通注释"), "debt"))
}

func TestCheck_AC_031_1_E1(t *testing.T) {
	dir := debtRepo(t, debtList("DEBT-001", "Open", "MOD-07"), taskList("Todo"), "")
	wantViolation(t, checkJSON(t, dir, "debt"), "debt.01", "DEBT-001", "MOD-07 不存在")
}

func TestCheck_AC_031_1_E2(t *testing.T) {
	dir := debtRepo(t, debtList("DEBT-001", "Done", "MOD-01"), taskList("Todo"), "")
	wantViolation(t, checkJSON(t, dir, "debt"), "debt.01", "DEBT-001", "Done", "Open、Resolved")
}

// ---- AC-031-2 待办注释引用 ----

func TestCheck_AC_031_2_N1(t *testing.T) {
	dir := debtRepo(t, debtList("DEBT-002", "Open", "MOD-01"), taskList("Todo"), "// TODO(DEBT-2): 重试")
	wantPass(t, checkJSON(t, dir, "debt"))
}

func TestCheck_AC_031_2_B1(t *testing.T) {
	// Release 1 只内置 Go 语言包，用 Go 注释写法验证"不带括号引用未完成任务"。
	dir := debtRepo(t, debtList(), taskList("InProgress"), "// FIXME T-5")
	wantPass(t, checkJSON(t, dir, "debt"))
}

func TestCheck_AC_031_2_E1(t *testing.T) {
	dir := debtRepo(t, debtList(), taskList("Todo"), "// TODO 以后再说")
	v := wantViolation(t, checkJSON(t, dir, "debt"), "debt.02", "TODO", "没有引用")
	if v.File != "internal/a.go" || v.Line != 3 {
		t.Fatalf("位置：%s:%d", v.File, v.Line)
	}
}

func TestCheck_AC_031_2_E2(t *testing.T) {
	dir := debtRepo(t, debtList(), taskList("Done"), "// TODO(T-5): 收尾")
	wantViolation(t, checkJSON(t, dir, "debt"), "debt.02", "T-005（Done）", "已关闭")
}

func TestCheck_AC_031_2_B2(t *testing.T) {
	dir := debtRepo(t, debtList(), taskList("Todo"), "var s = \"TODO\" // 字符串")
	wantPass(t, checkJSON(t, dir, "debt"))
}

// ---- AC-031-3 Resolved 的技术债不被引用 ----

func TestCheck_AC_031_3_N1(t *testing.T) {
	dir := debtRepo(t, debtList("DEBT-001", "Resolved", "MOD-01"), taskList("Todo"), "// 已修复")
	wantPass(t, checkJSON(t, dir, "debt"))
}

func TestCheck_AC_031_3_E1(t *testing.T) {
	dir := debtRepo(t, debtList("DEBT-001", "Resolved", "MOD-01"), taskList("Todo"), "// 绕过方式见 DEBT-001")
	wantViolation(t, checkJSON(t, dir, "debt"), "debt.03", "DEBT-001", "internal/a.go:3")
}
