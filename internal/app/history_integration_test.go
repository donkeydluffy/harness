package app_test

import (
	"strings"
	"testing"
)

// removedMod 是一个 Removed 的模块：Covers 为空，Depends 引用已 Removed 的需求，路径为 internal/old/**。
const removedMod = "\n### MOD-03 old\n- Status: Removed\n- Paths: internal/old/**\n- Public: internal/old/old.go\n- Depends: FR-005\n- Interfaces:\n- Covers:\n- Core: false\n- Coverage: 0\n"

// historyRepo 返回含 Removed 模块 MOD-03 与已 Removed 的 FR-005 的设计夹具。
func historyRepo(t *testing.T, archEdits ...string) map[string]string {
	t.Helper()
	src := prd(t, "## 4. 非功能需求", strings.Replace(strings.Replace(fr003R2, "FR-003", "FR-005", 1), "- Status: Active", "- Status: Removed", 1))
	files := traced(t, src, archEdits...)
	files["docs/architecture/overview.md"] += removedMod
	return files
}

// ---- AC-007-5 Removed 的设计元素与 Done 的任务是历史记录 ----

func TestCheck_AC_007_5_N1(t *testing.T) {
	wantPass(t, checkJSON(t, writeTree(t, historyRepo(t)), "arch", "trace"))
}

func TestCheck_AC_007_5_E1(t *testing.T) {
	files := historyRepo(t, "overview.md", "- Depends: MOD-02\n", "- Depends: MOD-02, MOD-03\n")
	wantViolation(t, checkJSON(t, writeTree(t, files), "trace"), "trace.03", "MOD-01", "已 Removed 的 MOD-03")
}

const historyTask = "# 任务列表\n\n- Status: Draft\n\n### T-001 某任务\n- Status: %s\n- Covers: FR-001\n- Design: %s\n- Scope: %s\n- Depends:\n"

func TestCheck_AC_007_5_B1(t *testing.T) {
	files := historyRepo(t)
	files["docs/tasks.md"] = sprintf(historyTask, "Todo", "MOD-03", "internal/old/**")
	wantPass(t, checkJSON(t, writeTree(t, files), "task.03", "task.05"))
}

func TestCheck_AC_007_5_B2(t *testing.T) {
	task := sprintf(historyTask, "Done", "MOD-01, MOD-02, FLOW-01", "internal/api/**, internal/store/**") + "- Flow: FLOW-01\n"
	narrowed := []string{"flows.md", "- Steps: MOD-01 -> IF-001, MOD-01 -> IF-002\n", "- Steps: MOD-01 -> IF-001\n"}
	files := historyRepo(t, narrowed...)
	files["docs/tasks.md"] = task
	wantPass(t, checkJSON(t, writeTree(t, files), "task.03", "task.05"))

	files = historyRepo(t, narrowed...)
	files["docs/tasks.md"] = strings.Replace(task, "- Status: Done", "- Status: Todo", 1)
	wantViolation(t, checkJSON(t, writeTree(t, files), "task.03", "task.05"), "task.05", "MOD-02")
}

// ---- AC-030-3/B2 ADR 与 Done 的任务中只检查编号存在 ----

func TestDoc_AC_030_3_B2(t *testing.T) {
	prd := strings.Replace(basePRD, "- Current-Release: 1", "- Current-Release: 2", 1)
	prd = strings.Replace(prd, "### FR-002 澄清\n- Status: Active\n- Release: 1\n", "### FR-002 澄清\n- Status: Active\n- Release: 1\n- Replaced-By: FR-045\n", 1)
	prd = strings.Replace(prd, "## 4. 非功能需求", "### FR-045 新澄清\n- Status: Active\n- Release: 2\n- Priority: P1\n- Replaces: FR-002\n\n- AC-045-1: 澄清。\n  - AC-045-1/N1: 澄清\n\n## 4. 非功能需求", 1)
	tasks := "# 任务列表\n\n- Status: Draft\n\n### T-001 旧任务\n- Status: Done\n- Covers: FR-045\n- Design:\n- Scope: internal/api/**\n- Depends:\n\n当时按 FR-002 实现。\n"
	adrText := sprintf(adr, "Accepted", "FR-045", "") + "\n当时按 FR-002 决定。\n"
	files := map[string]string{"docs/PRD.md": prd, "docs/tasks.md": tasks, "docs/adr/ADR-001.md": adrText}
	wantPass(t, checkDocs(t, docRepo(t, files)))

	files["docs/adr/ADR-001.md"] = adrText + "\n接口见 IF-099。\n"
	wantViolation(t, checkDocs(t, docRepo(t, files)), "doc.04", "IF-099", "不存在")
}

// adrRemoving 是决定删除扩展点 ext 的 ADR。
func adrRemoving(ext string) string {
	return "# ADR-001 删除旧扩展点\n\n- Status: Proposed\n- Relates: FR-001, " + ext +
		"\n\n## 背景\n\n旧扩展点不再需要。\n\n## 决策\n\n删除它。\n\n## 后果\n\n不再支持。\n"
}

func TestCheck_AC_007_5_B3(t *testing.T) {
	files := historyRepo(t)
	files["docs/architecture/overview.md"] += "\n### EXT-09 旧扩展点\n- Status: Removed\n- Interface: IF-001\n- Paths: internal/old/**\n- Closed: MOD-01\n- Relates: NFR-001\n"
	files["docs/adr/ADR-001-remove-ext.md"] = adrRemoving("EXT-09")
	wantPass(t, checkJSON(t, writeTree(t, files), "trace", "adr"))
	files["docs/adr/ADR-001-remove-ext.md"] = adrRemoving("EXT-99")
	wantViolation(t, checkJSON(t, writeTree(t, files), "trace"), "trace.03", "ADR-001", "不存在的 EXT-99")
}
