package app_test

import (
	"strings"
	"testing"
)

// task 生成一个任务；extra 为附加字段行（如 "- Flow: FLOW-01\n"）。
func task(id, status, covers, design, scope, depends, extra string) string {
	return "\n### " + id + " 某任务\n- Status: " + status + "\n- Covers: " + covers + "\n- Design: " + design +
		"\n- Scope: " + scope + "\n- Depends: " + depends + "\n" + extra
}

func taskDoc(ts ...string) string {
	return "# 任务列表\n\n- Status: Draft\n" + strings.Join(ts, "")
}

var (
	t001 = task("T-001", "Done", "FR-001", "MOD-01", "internal/api/**", "", "")
	t002 = task("T-002", "InProgress", "FR-002", "MOD-02", "internal/store/**", "T-001", "")
)

// taskRepo 写入 PRD、架构夹具与任务列表（tasks 为空串时不写任务列表），运行 check task。
func checkTasks(t *testing.T, prdText, tasks string) checked {
	t.Helper()
	files := arch(t)
	files["docs/PRD.md"] = prdText
	if tasks != "" {
		files["docs/tasks.md"] = tasks
	}
	return checkJSON(t, writeTree(t, files), taskRules...)
}

// task.07、task.08 依赖批准与测试报告，由 dod_integration_test.go 单独测试。
var taskRules = []string{"task.01", "task.02", "task.03", "task.04", "task.05", "task.06"}

// ---- AC-011-1 任务必填字段 ----

func TestCheck_AC_011_1_N1(t *testing.T) {
	wantPass(t, checkTasks(t, basePRD, taskDoc(t001, t002)))
}

func TestCheck_AC_011_1_N2(t *testing.T) {
	t2 := task("T-002", "InProgress", "FR-002, AC-001-1", "MOD-02", "internal/store/**", "T-001", "")
	wantPass(t, checkTasks(t, basePRD, taskDoc(t001, t2)))
}

func TestCheck_AC_011_1_E1(t *testing.T) {
	t2 := task("T-002", "InProgress", "FR-002", "", "internal/store/**", "T-001", "")
	wantViolation(t, checkTasks(t, basePRD, taskDoc(t001, t2)), "task.01", "T-002", "缺少 Design")
}

func TestCheck_AC_011_1_E2(t *testing.T) {
	t2 := task("T-002", "Doing", "FR-002", "MOD-02", "internal/store/**", "T-001", "")
	wantViolation(t, checkTasks(t, basePRD, taskDoc(t001, t2)), "task.01", "Doing", "Todo、InProgress、Done、Blocked")
}

func TestCheck_AC_011_1_B1(t *testing.T) {
	wantPass(t, checkTasks(t, basePRD, taskDoc(t001, t002))) // T-001 的 Depends 为空
}

func TestCheck_AC_011_1_E3(t *testing.T) {
	t2 := task("T-002", "InProgress", "FR-002, AC-001-9", "MOD-02", "internal/store/**", "T-001", "")
	wantViolation(t, checkTasks(t, basePRD, taskDoc(t001, t2)), "task.01", "AC-001-9 不存在")
}

// ---- AC-011-2 FR 被任务覆盖 ----

// twoACs 让 FR-001 有两条验收标准。
func twoACs(t *testing.T) string {
	return prd(t, "  - AC-001-1/E1: 只读目录\n", "  - AC-001-1/E1: 只读目录\n- AC-001-2: 写入配置。\n  - AC-001-2/N1: 默认配置\n  - AC-001-2/E1: 配置已存在\n")
}

func TestCheck_AC_011_2_N1(t *testing.T) {
	wantPass(t, checkTasks(t, basePRD, taskDoc(t001, t002)))
}

func TestCheck_AC_011_2_E1(t *testing.T) {
	t2 := task("T-002", "InProgress", "NFR-001", "MOD-02", "internal/store/**", "T-001", "")
	wantViolation(t, checkTasks(t, basePRD, taskDoc(t001, t2)), "task.02", "FR-002", "没有被任务覆盖")
}

func TestCheck_AC_011_2_B1(t *testing.T) {
	wantPass(t, checkTasks(t, basePRD, ""))
}

func TestCheck_AC_011_2_B2(t *testing.T) {
	t1 := task("T-001", "Done", "AC-001-1", "MOD-01", "internal/api/a.go", "", "")
	t3 := task("T-003", "Todo", "AC-001-2", "MOD-01", "internal/api/b.go", "T-002", "")
	wantPass(t, checkTasks(t, twoACs(t), taskDoc(t1, t002, t3)))
}

func TestCheck_AC_011_2_E2(t *testing.T) {
	t1 := task("T-001", "Done", "AC-001-1", "MOD-01", "internal/api/a.go", "", "")
	v := wantViolation(t, checkTasks(t, twoACs(t), taskDoc(t1, t002)), "task.02", "FR-001", "AC-001-2")
	if strings.Contains(v.Msg, "AC-001-1") {
		t.Fatalf("已关联的 AC-001-1 不应列出：%s", v.Msg)
	}
}

// ---- AC-011-3 范围落在关联模块内 ----

func TestCheck_AC_011_3_N1(t *testing.T) {
	t1 := task("T-001", "Done", "FR-001", "MOD-01", "internal/api/handler/**", "", "")
	wantPass(t, checkTasks(t, basePRD, taskDoc(t1, t002)))
}

func TestCheck_AC_011_3_E1(t *testing.T) {
	t1 := task("T-001", "Done", "FR-001", "MOD-01", "internal/api/**, internal/gate/x.go", "", "")
	wantViolation(t, checkTasks(t, basePRD, taskDoc(t1, t002)), "task.03", "T-001", "internal/gate/x.go")
	// 路径属于存在但未关联的模块 MOD-02，同样失败。
	t1 = task("T-001", "Done", "FR-001", "MOD-01", "internal/api/**, internal/store/x.go", "", "")
	wantViolation(t, checkTasks(t, basePRD, taskDoc(t1, t002)), "task.03", "T-001", "internal/store/x.go")
}

func TestCheck_AC_011_3_B1(t *testing.T) {
	wantPass(t, checkTasks(t, basePRD, taskDoc(t001, t002))) // 范围与 MOD-01 的 Paths 相同
}

// ---- AC-011-4 InProgress 唯一；重叠须有前置关系 ----

func TestCheck_AC_011_4_N1(t *testing.T) {
	wantPass(t, checkTasks(t, basePRD, taskDoc(t001, t002)))
}

func TestCheck_AC_011_4_E1(t *testing.T) {
	t1 := strings.Replace(t001, "Done", "InProgress", 1)
	wantViolation(t, checkTasks(t, basePRD, taskDoc(t1, t002)), "task.04", "2 个 InProgress", "T-001, T-002")
}

func TestCheck_AC_011_4_E2(t *testing.T) {
	t3 := task("T-003", "Todo", "FR-002", "MOD-02", "internal/store/a.go", "", "")
	t4 := task("T-004", "Todo", "FR-002", "MOD-02", "internal/store/**", "", "")
	c := checkTasks(t, basePRD, taskDoc(t001, strings.Replace(t002, "InProgress", "Done", 1), t3, t4))
	wantViolation(t, c, "task.04", "T-003 与 T-004", "internal/store/a.go 与 internal/store/**")
}

func TestCheck_AC_011_4_B1(t *testing.T) {
	t3 := task("T-003", "Todo", "FR-002", "MOD-02", "internal/store/a.go", "T-002", "")
	wantPass(t, checkTasks(t, basePRD, taskDoc(t001, t002, t3)))
}

func TestCheck_AC_011_4_B2(t *testing.T) {
	t3 := task("T-003", "Todo", "FR-001", "MOD-01", "internal/api/a.go", "", "")
	wantPass(t, checkTasks(t, basePRD, taskDoc(t001, t002, t3))) // 与 Done 的 T-001 重叠
}

// ---- AC-011-6 单模块或引用流程 ----

func TestCheck_AC_011_6_N1(t *testing.T) {
	wantPass(t, checkTasks(t, basePRD, taskDoc(t001, t002)))
}

func TestCheck_AC_011_6_E1(t *testing.T) {
	t2 := task("T-002", "InProgress", "FR-002", "MOD-01, MOD-02", "internal/store/**, internal/api/x.go", "T-001", "")
	wantViolation(t, checkTasks(t, basePRD, taskDoc(t001, t2)), "task.05", "T-002", "跨 2 个模块", "没有引用调用流程")
}

func TestCheck_AC_011_6_B1(t *testing.T) {
	t2 := task("T-002", "InProgress", "FR-002", "FLOW-01", "internal/store/**, internal/api/x.go", "T-001", "- Flow: FLOW-01\n")
	wantPass(t, checkTasks(t, basePRD, taskDoc(t001, t2)))
}

func TestCheck_AC_011_6_E2(t *testing.T) {
	files := arch(t, "overview.md", "\n### EXT-01", "\n### MOD-03 cache\n- Status: Active\n- Paths: internal/cache/**\n- Public: internal/cache/cache.go\n- Depends:\n- Interfaces:\n- Covers: FR-002\n- Core: false\n- Coverage: 60\n\n### EXT-01")
	files["docs/PRD.md"] = basePRD
	files["docs/tasks.md"] = taskDoc(t001, task("T-002", "InProgress", "FR-002", "FLOW-01, MOD-03", "internal/store/**, internal/cache/**", "T-001", "- Flow: FLOW-01\n"))
	wantViolation(t, checkJSON(t, writeTree(t, files), taskRules...), "task.05", "MOD-03", "不在引用的流程 FLOW-01")
}

// ---- AC-034-2 扩展任务不改封闭模块 ----

func TestCheck_AC_034_2_N1(t *testing.T) {
	t3 := task("T-003", "Todo", "NFR-002", "MOD-02", "internal/store/mem/**", "T-002", "- Extends: EXT-01\n")
	wantPass(t, checkTasks(t, basePRD, taskDoc(t001, t002, t3)))
}

func TestCheck_AC_034_2_E1(t *testing.T) {
	t3 := task("T-003", "Todo", "NFR-002", "FLOW-01", "internal/store/mem/**, internal/api/mem.go", "T-002", "- Extends: EXT-01\n- Flow: FLOW-01\n")
	wantViolation(t, checkTasks(t, basePRD, taskDoc(t001, t002, t3)), "task.06", "T-003", "封闭模块 MOD-01", "internal/api/mem.go")
}
