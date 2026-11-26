package app_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// unitRepo 返回锁文件为第 2 版（按单位批准）的仓库，PRD 与设计尚未批准。
func unitRepo(t *testing.T) string {
	t.Helper()
	dir := lockedRepo(t, basePRD)
	p := filepath.Join(dir, "harness.lock")
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), `"version": 1`) {
		t.Fatalf("锁文件版本：%s", b)
	}
	os.WriteFile(p, []byte(strings.Replace(string(b), `"version": 1`, `"version": 2`, 1)), 0o644)
	return dir
}

// approveUnit 在伪终端中执行 harness approve target 并输入 answer。
func approveUnit(t *testing.T, dir, target, answer string) (string, int) {
	t.Helper()
	return inTerminal(t, dir, confirmPrompt, answer, "approve", target)
}

func mustApprove(t *testing.T, dir string, targets ...string) {
	t.Helper()
	for _, target := range targets {
		if out, code := approveUnit(t, dir, target, "y"); code != 0 {
			t.Fatalf("批准 %s 失败：%d %s", target, code, out)
		}
	}
}

// unitRecords 返回锁文件中按单位的批准记录。
func unitRecords(t *testing.T, dir string) map[string]map[string]any {
	t.Helper()
	var l struct{ Units map[string]map[string]any }
	json.Unmarshal([]byte(readFile(t, dir, "harness.lock")), &l)
	return l.Units
}

// approvalWarnings 返回 approve.01 的告警文字。
func approvalWarnings(t *testing.T, dir string) string {
	t.Helper()
	r := runHarness(t, dir, "check", "approve", "--json")
	var rep struct{ Warnings []violation }
	if err := json.Unmarshal([]byte(r.Stdout), &rep); err != nil {
		t.Fatalf("解析输出：%v\n%s", err, r.Stdout)
	}
	var b strings.Builder
	for _, w := range rep.Warnings {
		b.WriteString(w.Msg + "\n")
	}
	if r.Code != 0 {
		t.Fatalf("告警不应影响退出码：%d %s", r.Code, r.Stdout)
	}
	return b.String()
}

func editFile(t *testing.T, dir, rel string, pairs ...string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	b, _ := os.ReadFile(p)
	s := string(b)
	for i := 0; i+1 < len(pairs); i += 2 {
		if !strings.Contains(s, pairs[i]) {
			t.Fatalf("%s 中没有 %q", rel, pairs[i])
		}
		s = strings.Replace(s, pairs[i], pairs[i+1], 1)
	}
	os.WriteFile(p, []byte(s), 0o644)
}

const fr003 = "### FR-003 导出\n- Status: Active\n- Release: 1\n- Priority: P1\n\n- AC-003-1: 导出报告。\n  - AC-003-1/N1: 导出一份\n  - AC-003-1/E1: 目录不可写\n\n## 4. 非功能需求"

// addFR003 新增 FR-003 并登记变更。
func addFR003(t *testing.T, dir string) {
	editFile(t, dir, "docs/PRD.md", "## 4. 非功能需求", fr003,
		"| CHG-001 | 2026-11-25 | 新增 | NFR-001, NFR-002 | 初稿 |", "| CHG-001 | 2026-11-25 | 新增 | NFR-001, NFR-002 | 初稿 |\n| CHG-002 | 2026-11-26 | 新增 | FR-003 | 导出 |")
}

// ---- AC-047-1 按单位记录批准 ----

func TestApprove_AC_047_1_N1(t *testing.T) {
	dir := unitRepo(t)
	mustApprove(t, dir, "prd")
	editFile(t, dir, "docs/PRD.md", "- AC-002-1: 记录问答。", "- AC-002-1: 记录问答与结论。")
	w := approvalWarnings(t, dir)
	if !strings.Contains(w, "FR-002 批准已失效") || strings.Contains(w, "FR-001") || strings.Contains(w, "NFR-001") {
		t.Fatalf("只有 FR-002 应失效：\n%s", w)
	}
}

func TestApprove_AC_047_1_N2(t *testing.T) {
	dir := unitRepo(t)
	mustApprove(t, dir, "prd")
	editFile(t, dir, "docs/PRD.md", "- Current-Release: 1\n", "- Current-Release: 1\n\n背景：补充一段说明。\n")
	w := approvalWarnings(t, dir)
	if !strings.Contains(w, "docs/PRD.md#正文 批准已失效") || strings.Contains(w, "FR-0") {
		t.Fatalf("只有正文应失效：\n%s", w)
	}
}

func TestApprove_AC_047_1_B1(t *testing.T) {
	dir := unitRepo(t)
	mustApprove(t, dir, "prd")
	editFile(t, dir, "docs/PRD.md", "- AC-001-1: 生成骨架。\n", "- AC-001-1: 生成骨架。  \n", "## 3. 功能需求\n", "## 3. 功能需求\t\n\n\n")
	if w := approvalWarnings(t, dir); strings.Contains(w, "docs/PRD.md") || strings.Contains(w, "FR-001") {
		t.Fatalf("只改空白不应失效：\n%s", w)
	}
}

func TestApprove_AC_047_1_E1(t *testing.T) {
	dir := unitRepo(t)
	mustApprove(t, dir, "prd")
	addFR003(t, dir)
	if w := approvalWarnings(t, dir); !strings.Contains(w, "FR-003 未批准") || strings.Contains(w, "FR-001") {
		t.Fatalf("新增的 FR-003 应列为未批准：\n%s", w)
	}
}

// ---- AC-047-2 harness approve 按单位批准 ----

func TestApprove_AC_047_2_N1(t *testing.T) {
	dir := unitRepo(t)
	mustApprove(t, dir, "prd")
	addFR003(t, dir)
	editFile(t, dir, "docs/PRD.md", "- AC-001-1: 生成骨架。", "- AC-001-1: 生成骨架与配置。",
		"| CHG-002 | 2026-11-26 | 新增 | FR-003 | 导出 |", "| CHG-002 | 2026-11-26 | 新增 | FR-003 | 导出 |\n| CHG-002 | 2026-11-26 | 修改 | FR-001 | 生成配置 |")
	out, code := approveUnit(t, dir, "prd", "y")
	for _, want := range []string{"新增 FR-003", "修改 FR-001", "- - AC-001-1: 生成骨架。", "+ - AC-001-1: 生成骨架与配置。", "批准以上 2 个单位"} {
		if !strings.Contains(out, want) {
			t.Errorf("输出应包含 %q：\n%s", want, out)
		}
	}
	if code != 0 {
		t.Fatalf("退出码 %d", code)
	}
	u := unitRecords(t, dir)
	if u["FR-003"]["by"] == "" || u["FR-001"]["hash"] == "" {
		t.Fatalf("批准记录：%v", u)
	}
	if w := approvalWarnings(t, dir); strings.Contains(w, "FR-00") {
		t.Fatalf("批准后不应有 PRD 需求的告警：\n%s", w)
	}
}

func TestApprove_AC_047_2_N2(t *testing.T) {
	dir := unitRepo(t)
	addFR003(t, dir)
	mustApprove(t, dir, "FR-003")
	u := unitRecords(t, dir)
	if len(u) != 1 || u["FR-003"] == nil {
		t.Fatalf("只应批准 FR-003：%v", u)
	}
}

func TestApprove_AC_047_2_E1(t *testing.T) {
	dir := unitRepo(t)
	editFile(t, dir, "docs/PRD.md", "  - AC-002-1/B1: 一轮 5 个问题\n", "")
	before := readFile(t, dir, "harness.lock")
	out, code := approveUnit(t, dir, "prd", "y")
	if code != 1 || !strings.Contains(out, "prd.07") || !strings.Contains(out, "拒绝批准") {
		t.Fatalf("FR-002 检查失败时应拒绝：%d %s", code, out)
	}
	if readFile(t, dir, "harness.lock") != before {
		t.Fatal("拒绝后锁文件不应改变")
	}
}

func TestApprove_AC_047_2_B1(t *testing.T) {
	dir := unitRepo(t)
	mustApprove(t, dir, "prd")
	before := readFile(t, dir, "harness.lock")
	out, code := inTerminal(t, dir, "", "", "approve", "prd")
	if code != 0 || !strings.Contains(out, "没有待批准的内容") {
		t.Fatalf("没有待批准单位时应提示并退出 0：%d %s", code, out)
	}
	if readFile(t, dir, "harness.lock") != before {
		t.Fatal("锁文件不应改变")
	}
}

// ---- AC-047-3 批准链 ----

func TestApprove_AC_047_3_N1(t *testing.T) {
	dir := unitRepo(t)
	mustApprove(t, dir, "prd", "MOD-02")
	if u := unitRecords(t, dir)["MOD-02"]; u == nil {
		t.Fatal("MOD-02 应已批准")
	}
}

func TestApprove_AC_047_3_E1(t *testing.T) {
	dir := unitRepo(t)
	mustApprove(t, dir, "prd")
	addFR003(t, dir)
	editFile(t, dir, "docs/architecture/overview.md", "- Covers: FR-002", "- Covers: FR-002, FR-003")
	out, code := approveUnit(t, dir, "MOD-02", "y")
	if code != 1 || !strings.Contains(out, "FR-003 未批准") {
		t.Fatalf("覆盖未批准的 FR-003 时应拒绝：%d %s", code, out)
	}
	if unitRecords(t, dir)["MOD-02"] != nil {
		t.Fatal("MOD-02 不应被批准")
	}
}

func TestApprove_AC_047_3_E2(t *testing.T) {
	dir := unitRepo(t)
	mustApprove(t, dir, "prd", "MOD-02")
	editFile(t, dir, "docs/PRD.md", "- AC-002-1: 记录问答。", "- AC-002-1: 记录问答与结论。")
	if w := approvalWarnings(t, dir); !strings.Contains(w, "MOD-02 批准已失效：FR-002") {
		t.Fatalf("MOD-02 应因 FR-002 变化失效：\n%s", w)
	}
}

func TestApprove_AC_047_3_B1(t *testing.T) {
	dir := unitRepo(t)
	mustApprove(t, dir, "prd", "MOD-02")
	editFile(t, dir, "docs/PRD.md", "- AC-002-1: 记录问答。\n", "- AC-002-1: 记录问答。   \n")
	if w := approvalWarnings(t, dir); strings.Contains(w, "MOD-02") {
		t.Fatalf("FR-002 只改空白，MOD-02 不应失效：\n%s", w)
	}
}
