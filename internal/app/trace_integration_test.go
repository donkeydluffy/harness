package app_test

import (
	"encoding/json"
	"strings"
	"testing"
)

// traced 返回 PRD 与架构的组合夹具；prdPairs 作用于 PRD，archEdits 同 arch()。
func traced(t *testing.T, prdText string, archEdits ...string) map[string]string {
	t.Helper()
	files := arch(t, archEdits...)
	files["docs/PRD.md"] = prdText
	return files
}

const adr = "# ADR-001 存储选型\n\n- Status: %s\n- Relates: %s\n%s\n## 背景\n\n文字\n\n## 决策\n\n文字\n\n## 后果\n\n文字\n"

func sprintf(f string, a ...any) string {
	for _, x := range a {
		f = strings.Replace(f, "%s", x.(string), 1)
	}
	return f
}

func checkTrace(t *testing.T, files map[string]string) checked {
	t.Helper()
	return checkJSON(t, writeTree(t, files), "trace")
}

const fr003R2 = "### FR-003 下一期\n- Status: Active\n- Release: 2\n- Priority: P2\n\n- AC-003-1: 下一期。\n  - AC-003-1/N1: 正常\n  - AC-003-1/E1: 异常\n\n## 4. 非功能需求"

// ---- AC-007-1 FR 被设计覆盖 ----

func TestCheck_AC_007_1_N1(t *testing.T) { wantPass(t, checkTrace(t, traced(t, basePRD))) }

func TestCheck_AC_007_1_E1(t *testing.T) {
	src := prd(t, "## 4. 非功能需求", strings.Replace(fr003R2, "- Release: 2", "- Release: 1", 1))
	wantViolation(t, checkTrace(t, traced(t, src)), "trace.01", "FR-003 未被任何设计元素覆盖")
}

func TestCheck_AC_007_1_B1(t *testing.T) {
	src := prd(t, "## 4. 非功能需求", strings.Replace(fr003R2, "- Release: 2", "- Release: 1", 1))
	files := traced(t, src)
	files["docs/adr/ADR-001.md"] = sprintf(adr, "Superseded", "FR-003", "- Superseded-By: ADR-002\n")
	files["docs/adr/ADR-002.md"] = strings.Replace(sprintf(adr, "Accepted", "FR-001", ""), "ADR-001", "ADR-002", 1)
	wantViolation(t, checkTrace(t, files), "trace.01", "FR-003")
}

func TestCheck_AC_007_1_B2(t *testing.T) {
	src := prd(t, "## 4. 非功能需求", strings.Replace(strings.Replace(fr003R2, "- Release: 2", "- Release: 1", 1), "- Status: Active", "- Status: Removed", 1))
	wantPass(t, checkTrace(t, traced(t, src)))
}

// ---- AC-007-2 孤儿 ----

func TestCheck_AC_007_2_N1(t *testing.T) { wantPass(t, checkTrace(t, traced(t, basePRD))) }

func TestCheck_AC_007_2_E1(t *testing.T) {
	c := checkTrace(t, traced(t, basePRD, "overview.md", "- Covers: FR-002\n", "- Covers:\n"))
	wantViolation(t, c, "trace.02", "MOD-02", "孤儿")
}

// ---- AC-007-3 悬空引用 ----

func TestCheck_AC_007_3_N1(t *testing.T) { wantPass(t, checkTrace(t, traced(t, basePRD))) }

func TestCheck_AC_007_3_E1(t *testing.T) {
	c := checkTrace(t, traced(t, basePRD, "overview.md", "- Covers: FR-002\n", "- Covers: FR-002, FR-099\n"))
	wantViolation(t, c, "trace.03", "MOD-02", "不存在的 FR-099")
}

func TestCheck_AC_007_3_E2(t *testing.T) {
	src := prd(t, "## 4. 非功能需求", strings.Replace(strings.Replace(fr003R2, "FR-003", "FR-005", 1), "- Status: Active", "- Status: Removed", 1))
	c := checkTrace(t, traced(t, src, "overview.md", "- Covers: FR-002\n", "- Covers: FR-002, FR-005\n"))
	wantViolation(t, c, "trace.03", "已 Removed 的 FR-005")
}

func TestCheck_AC_007_3_B1(t *testing.T) {
	files := traced(t, basePRD, "overview.md", "- Relates: NFR-001\n", "- Relates: NFR-001, ADR-001\n")
	files["docs/adr/ADR-001.md"] = sprintf(adr, "Superseded", "FR-001", "- Superseded-By: ADR-002\n")
	files["docs/adr/ADR-002.md"] = strings.Replace(sprintf(adr, "Accepted", "FR-001", ""), "ADR-001", "ADR-002", 1)
	wantViolation(t, checkTrace(t, files), "trace.03", "Superseded 的 ADR-001", "应改为引用 ADR-002")
}

// ---- AC-007-4 覆盖矩阵 ----

func TestTrace_AC_007_4_N1(t *testing.T) {
	r := runHarness(t, writeTree(t, traced(t, basePRD)), "trace", "--format", "md")
	lines := strings.Split(strings.TrimSpace(r.Stdout), "\n")
	if r.Code != 0 || len(lines) != 2+3 { // 表头两行 + FR-001、FR-002、NFR-001
		t.Fatalf("退出码 %d，输出：\n%s", r.Code, r.Stdout)
	}
	if !strings.Contains(r.Stdout, "| FR-001 | 初始化 | FLOW-01, IF-001, MOD-01 |") {
		t.Fatalf("矩阵内容：\n%s", r.Stdout)
	}
}

func TestTrace_AC_007_4_N2(t *testing.T) {
	dir := writeTree(t, traced(t, basePRD))
	md := runHarness(t, dir, "trace", "--format", "md")
	js := runHarness(t, dir, "trace", "--format", "json")
	var out struct {
		Rows []struct {
			Requirement, Title string
			Elements           []string
		}
	}
	if err := json.Unmarshal([]byte(js.Stdout), &out); err != nil || js.Code != 0 {
		t.Fatalf("JSON 解析失败：%v\n%s", err, js.Stdout)
	}
	var b strings.Builder
	b.WriteString("| 需求 | 标题 | 设计元素 |\n|---|---|---|\n")
	for _, r := range out.Rows {
		b.WriteString("| " + r.Requirement + " | " + r.Title + " | " + strings.Join(r.Elements, ", ") + " |\n")
	}
	if b.String() != md.Stdout {
		t.Fatalf("JSON 与 Markdown 不一致：\n%s\n%s", b.String(), md.Stdout)
	}
}

func TestTrace_AC_007_4_B1(t *testing.T) {
	r := runHarness(t, writeTree(t, map[string]string{"docs/PRD.md": basePRD}), "trace", "--format", "json")
	if r.Code != 0 || !strings.Contains(r.Stdout, `"elements": []`) {
		t.Fatalf("退出码 %d，输出：\n%s", r.Code, r.Stdout)
	}
}

func TestTrace_AC_007_4_E1(t *testing.T) {
	r := runHarness(t, writeTree(t, traced(t, basePRD)), "trace", "--format", "xml")
	if r.Code == 0 || !strings.Contains(r.Stderr, "md, json") {
		t.Fatalf("退出码 %d：%s", r.Code, r.Stderr)
	}
}

// ---- AC-003-5 Release 范围 ----

func TestCheck_AC_003_5_N1(t *testing.T) {
	wantPass(t, checkTrace(t, traced(t, prd(t, "## 4. 非功能需求", fr003R2))))
}

func TestCheck_AC_003_5_B1(t *testing.T) {
	src := prd(t, "## 4. 非功能需求", fr003R2, "- Current-Release: 1", "- Current-Release: 2")
	wantViolation(t, checkTrace(t, traced(t, src)), "trace.01", "FR-003")
}

func TestCheck_AC_003_5_E1(t *testing.T) {
	src := prd(t, "### FR-002 澄清\n- Status: Active\n- Release: 1\n", "### FR-002 澄清\n- Status: Active\n")
	wantViolation(t, checkJSON(t, writeTree(t, traced(t, src))), "prd.02", "FR-002 缺少 Release")
}

func TestCheck_AC_003_5_E2(t *testing.T) {
	src := prd(t, "- Current-Release: 1\n", "")
	wantViolation(t, checkJSON(t, writeTree(t, traced(t, src))), "prd.02", "Current-Release")
}
