package app_test

import (
	"strings"
	"testing"
)

// tasksWithAlias 生成第 lineNo 行含指定文字的任务列表。
func tasksWithAlias(lineNo int, text string) string {
	lines := []string{"# 任务列表", "", "- Status: Draft"}
	for len(lines) < lineNo-1 {
		lines = append(lines, "")
	}
	lines = append(lines, text)
	return strings.Join(lines, "\n") + "\n"
}

// archWithDomain 生成带 Domain 字段的架构声明。
func archWithDomain(domain string) string {
	return "# 架构总览\n\n- Status: Draft\n\n### MOD-01 order\n- Status: Active\n- Domain: " + domain + "\n\n订单模块，内部使用 `HttpServer`。\n"
}

func termPRD(t *testing.T) string {
	return prd(t, "| TERM-001 | 产物 | Artifact | 结构化文档 | 交付物 |",
		"| TERM-001 | 产物 | Artifact | 结构化文档 | 交付物 |\n| TERM-002 | 任务 | Task | 任务列表中的一项 | 工单 |\n| TERM-003 | 订单 | Order | 客户下的单 | |")
}

// ---- AC-005-1 禁用别名 ----

func TestCheck_AC_005_1_N1(t *testing.T) {
	dir := writeTree(t, map[string]string{
		"docs/PRD.md":   termPRD(t),
		"docs/tasks.md": tasksWithAlias(12, "实现任务 T-001。"),
	})
	wantPass(t, checkJSON(t, dir, "term"))
}

func TestCheck_AC_005_1_E1(t *testing.T) {
	dir := writeTree(t, map[string]string{
		"docs/PRD.md":   termPRD(t),
		"docs/tasks.md": tasksWithAlias(12, "这个工单要先做。"),
	})
	v := wantViolation(t, checkJSON(t, dir, "term"), "term.01", "工单", "任务")
	if v.File != "docs/tasks.md" || v.Line != 12 {
		t.Fatalf("位置：%s:%d", v.File, v.Line)
	}
}

func TestCheck_AC_005_1_B1(t *testing.T) {
	dir := writeTree(t, map[string]string{
		"docs/PRD.md":   termPRD(t),
		"docs/tasks.md": tasksWithAlias(12, "旧系统把它叫 `工单`。\n\n```\n工单\n```"),
	})
	wantPass(t, checkJSON(t, dir, "term"))
}

func TestCheck_AC_005_1_B2(t *testing.T) {
	// 统一语言表格本身的禁用别名列含"交付物""工单"，不报告。
	wantPass(t, checkJSON(t, writeTree(t, map[string]string{"docs/PRD.md": termPRD(t)}), "term"))
}

// ---- AC-005-2 领域概念在术语表中 ----

func TestCheck_AC_005_2_N1(t *testing.T) {
	dir := writeTree(t, map[string]string{"docs/PRD.md": termPRD(t), "docs/architecture/overview.md": archWithDomain("Order")})
	wantPass(t, checkJSON(t, dir, "term"))
}

func TestCheck_AC_005_2_E1(t *testing.T) {
	dir := writeTree(t, map[string]string{"docs/PRD.md": termPRD(t), "docs/architecture/overview.md": archWithDomain("Order, Ticket")})
	wantViolation(t, checkJSON(t, dir, "term"), "term.02", "Ticket")
}

func TestCheck_AC_005_2_B1(t *testing.T) {
	// HttpServer 没有用 Domain 标注为领域概念，不在术语表中也不报告。
	dir := writeTree(t, map[string]string{"docs/PRD.md": termPRD(t), "docs/architecture/overview.md": archWithDomain("Order")})
	wantPass(t, checkJSON(t, dir, "term"))
}
