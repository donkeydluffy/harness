package app_test

import (
	"strings"
	"testing"
)

// basePRD 中 NFR-002 扩展场景的完整文本。
const nfr002 = "### NFR-002 扩展场景：新增存储\n- Status: Active\n- Release: 2\n- Type: Scenario\n- Priority: P2\n- Adds: 一种存储实现\n- Allowed-Paths: 存储实现目录\n- Closed-Modules: 接口层\n\n- AC-N002-1: 新增一种存储只改存储实现目录。\n\n"

const ext02 = "\n### EXT-02 新增接口\n- Status: Active\n- Interface: IF-001\n- Paths: internal/api/<实现>/**\n- Closed: MOD-02\n"

// extRepo 写入 PRD、架构夹具与可选的测试报告。
func extRepo(t *testing.T, prdText, report string, archEdits ...string) string {
	t.Helper()
	files := arch(t, archEdits...)
	files["docs/PRD.md"] = prdText
	if report != "" {
		files[".harness/reports/go.jsonl"] = report
	}
	return writeTree(t, files)
}

// ---- AC-033-1 扩展场景字段 ----

func TestCheck_AC_033_1_N1(t *testing.T) {
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": basePRD}))
}

func TestCheck_AC_033_1_E1(t *testing.T) {
	c := checkPRD(t, map[string]string{"docs/PRD.md": prd(t, "- Closed-Modules: 接口层\n", "")})
	wantViolation(t, c, "prd.11", "NFR-002", "Closed-Modules")
}

// ---- AC-033-3 扩展点须关联扩展场景 ----

func TestCheck_AC_033_3_N1(t *testing.T) {
	wantPass(t, checkJSON(t, extRepo(t, basePRD, ""), "arch"))
}

func TestCheck_AC_033_3_E1(t *testing.T) {
	dir := extRepo(t, basePRD, "", "overview.md", "- Relates: NFR-002\n", "- Relates: NFR-002\n"+ext02)
	wantViolation(t, checkJSON(t, dir, "arch"), "arch.04", "EXT-02", "过度设计")
}

func TestCheck_AC_033_3_E2(t *testing.T) {
	dir := extRepo(t, basePRD, "", "overview.md", "- Relates: NFR-002\n", "- Relates: NFR-002\n"+ext02+"- Relates: NFR-001\n")
	wantViolation(t, checkJSON(t, dir, "arch"), "arch.04", "EXT-02", "NFR-001 不是 Scenario")
}

// ---- AC-033-4 扩展点有通过的示例扩展测试 ----

func release1(t *testing.T) string {
	return prd(t, nfr002, strings.Replace(nfr002, "- Release: 2", "- Release: 1", 1))
}

func TestCheck_AC_033_4_N1(t *testing.T) {
	dir := extRepo(t, release1(t), goReport("TestExample_EXT_01", "pass"))
	wantPass(t, checkJSON(t, dir, "arch"))
}

func TestCheck_AC_033_4_E1(t *testing.T) {
	dir := extRepo(t, release1(t), goReport("TestOther", "pass", "TestExample_EXT_01", "fail"))
	wantViolation(t, checkJSON(t, dir, "arch"), "arch.05", "EXT-01", "EXT_01")
}

func TestCheck_AC_033_4_B1(t *testing.T) {
	wantPass(t, checkJSON(t, extRepo(t, basePRD, ""), "arch")) // NFR-002 为 Release 2，没有报告也不检查
}

// ---- AC-033-5 扩展场景或"不预留扩展"声明 ----

func TestCheck_AC_033_5_N1(t *testing.T) {
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": basePRD}))
}

func TestCheck_AC_033_5_B1(t *testing.T) {
	src := prd(t, nfr002, "本项目不预留扩展。\n\n", "NFR-001, NFR-002 |", "NFR-001 |")
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": src}))
}

func TestCheck_AC_033_5_E1(t *testing.T) {
	src := prd(t, nfr002, "", "NFR-001, NFR-002 |", "NFR-001 |")
	wantViolation(t, checkPRD(t, map[string]string{"docs/PRD.md": src}), "prd.11", "不预留扩展")
}
