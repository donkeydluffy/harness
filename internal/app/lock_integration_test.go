package app_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// lockedRepo 写入 PRD 并执行 harness lock，返回目录。
func lockedRepo(t *testing.T, prdText string) string {
	t.Helper()
	dir := writeTree(t, map[string]string{"docs/PRD.md": prdText})
	if r := runHarness(t, dir, "lock"); r.Code != 0 {
		t.Fatalf("harness lock 失败：%s%s", r.Stdout, r.Stderr)
	}
	return dir
}

// edit 对目录中的 PRD 依次做字符串替换；每个替换必须命中。
func edit(t *testing.T, dir string, pairs ...string) {
	t.Helper()
	p := filepath.Join(dir, "docs/PRD.md")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for i := 0; i+1 < len(pairs); i += 2 {
		if !strings.Contains(s, pairs[i]) {
			t.Fatalf("替换目标不存在：%q", pairs[i])
		}
		s = strings.Replace(s, pairs[i], pairs[i+1], 1)
	}
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

// addCHG 在变更日志末尾追加行。
func addCHG(t *testing.T, dir string, rows ...string) {
	t.Helper()
	p := filepath.Join(dir, "docs/PRD.md")
	b, _ := os.ReadFile(p)
	s := strings.TrimRight(string(b), "\n") + "\n" + strings.Join(rows, "\n") + "\n"
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

const (
	fr002Desc    = "- AC-002-1: 记录问答。"
	fr002Changed = "- AC-002-1: 记录每轮问答。"
	fr020        = "### FR-020 新需求\n- Status: Active\n- Release: 1\n- Priority: P2\n\n- AC-020-1: 新增。\n  - AC-020-1/N1: 正常\n  - AC-020-1/E1: 异常\n\n## 4. 非功能需求"
)

// ---- AC-004-1 变更必须登记 ----

func TestCheck_AC_004_1_N1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	edit(t, dir, fr002Desc, fr002Changed)
	addCHG(t, dir, "| CHG-002 | 2026-11-26 | 修改 | FR-002 | 措辞 |")
	wantPass(t, checkJSON(t, dir))
}

func TestCheck_AC_004_1_E1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	edit(t, dir, fr002Desc, fr002Changed)
	wantViolation(t, checkJSON(t, dir), "prd.08", "FR-002", "没有引用它的新 CHG")
}

func TestCheck_AC_004_1_B1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	edit(t, dir, "### FR-002 澄清\n", "### FR-002 澄清   \n")
	wantPass(t, checkJSON(t, dir))
}

func TestCheck_AC_004_1_E2(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	edit(t, dir, fr002Desc, fr002Changed)
	addCHG(t, dir, "| CHG-002 | 2026-11-26 | 修改 | FR-001 | 写错对象 |")
	c := checkJSON(t, dir)
	wantViolation(t, c, "prd.08", "FR-002", "没有引用它的新 CHG")
	wantViolation(t, c, "prd.08", "FR-001 内容未变")
}

// ---- AC-004-2 动作与差异一致 ----

func TestCheck_AC_004_2_N1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	edit(t, dir, "## 4. 非功能需求", fr020)
	addCHG(t, dir, "| CHG-002 | 2026-11-26 | 新增 | FR-020 | 新需求 |")
	wantPass(t, checkJSON(t, dir))
}

func TestCheck_AC_004_2_E1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	edit(t, dir, "### FR-002 澄清\n- Status: Active", "### FR-002 澄清\n- Status: Removed")
	addCHG(t, dir, "| CHG-002 | 2026-11-26 | 修改 | FR-002 | 实为删除 |")
	wantViolation(t, checkJSON(t, dir), "prd.08", "FR-002", "应为删除")
}

func TestCheck_AC_004_2_E2(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	addCHG(t, dir, "| CHG-002 | 2026-11-26 | 删除 | FR-001 | 未删 |")
	wantViolation(t, checkJSON(t, dir), "prd.08", "声明删除 FR-001", "内容未变")
}

func TestCheck_AC_004_2_B1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	edit(t, dir, "## 4. 非功能需求", fr020, fr002Desc, fr002Changed)
	addCHG(t, dir,
		"| CHG-002 | 2026-11-26 | 新增 | FR-020 | 新需求 |",
		"| CHG-002 | 2026-11-26 | 修改 | FR-002 | 措辞 |")
	wantPass(t, checkJSON(t, dir))
}

// ---- AC-004-3 harness lock ----

func readLock(t *testing.T, dir string) (map[string]string, []byte) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "harness.lock"))
	if err != nil {
		t.Fatal(err)
	}
	var l struct{ Items map[string]string }
	if err := json.Unmarshal(b, &l); err != nil {
		t.Fatal(err)
	}
	return l.Items, b
}

func TestLock_AC_004_3_N1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	before, _ := readLock(t, dir)
	edit(t, dir, fr002Desc, fr002Changed)
	addCHG(t, dir, "| CHG-002 | 2026-11-26 | 修改 | FR-002 | 措辞 |")
	if r := runHarness(t, dir, "lock"); r.Code != 0 || !strings.Contains(r.Stdout, "已更新 harness.lock") {
		t.Fatalf("lock：%d %s%s", r.Code, r.Stdout, r.Stderr)
	}
	after, _ := readLock(t, dir)
	if after["FR-002"] == before["FR-002"] || after["CHG-002"] == "" || after["FR-001"] != before["FR-001"] {
		t.Fatalf("锁文件未按当前哈希更新：%v", after)
	}
}

func TestLock_AC_004_3_E1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	_, before := readLock(t, dir)
	edit(t, dir, fr002Desc, fr002Changed)
	r := runHarness(t, dir, "lock")
	if r.Code != 1 || !strings.Contains(r.Stderr, "锁文件未更新") {
		t.Fatalf("期望退出码 1：%d %s", r.Code, r.Stderr)
	}
	if _, after := readLock(t, dir); !bytes.Equal(before, after) {
		t.Fatal("检查未通过时锁文件被修改")
	}
}

func TestLock_AC_004_3_B1(t *testing.T) {
	dir := writeTree(t, map[string]string{"docs/PRD.md": basePRD})
	if _, err := os.Stat(filepath.Join(dir, "harness.lock")); err == nil {
		t.Fatal("夹具不应有锁文件")
	}
	if r := runHarness(t, dir, "lock"); r.Code != 0 {
		t.Fatalf("lock：%s", r.Stderr)
	}
	items, _ := readLock(t, dir)
	for _, k := range []string{"FR-001", "FR-002", "NFR-001", "AC-001-1/E1", "CHG-001"} {
		if !strings.HasPrefix(items[k], "sha256:") {
			t.Errorf("锁文件缺少 %s：%v", k, items)
		}
	}
}

// ---- AC-003-3 Removed 编号保留、不复用 ----

func TestCheck_AC_003_3_N1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	edit(t, dir, "### FR-002 澄清\n- Status: Active", "### FR-002 澄清\n- Status: Removed")
	addCHG(t, dir, "| CHG-002 | 2026-11-26 | 删除 | FR-002 | 不再需要 |")
	wantPass(t, checkJSON(t, dir))
}

func TestCheck_AC_003_3_E1(t *testing.T) {
	src := prd(t, "## 4. 非功能需求", fr020)
	dir := lockedRepo(t, src)
	edit(t, dir, fr020, "## 4. 非功能需求")
	wantViolation(t, checkJSON(t, dir), "prd.03", "FR-020 从文档中消失")
}

func TestCheck_AC_003_3_E2(t *testing.T) {
	dir := lockedRepo(t, prd(t, "### FR-002 澄清\n- Status: Active", "### FR-002 澄清\n- Status: Removed"))
	edit(t, dir, "### FR-002 澄清\n- Status: Removed", "### FR-002 新的需求\n- Status: Active")
	addCHG(t, dir, "| CHG-002 | 2026-11-26 | 新增 | FR-002 | 复用编号 |")
	wantViolation(t, checkJSON(t, dir), "prd.03", "FR-002 已 Removed，编号不能复用")
}

// ---- AC-019-3 用例编号只增不删 ----

func TestCheck_AC_019_3_N1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	edit(t, dir, "  - AC-001-1/E1: 只读目录\n", "  - AC-001-1/E1: 只读目录\n  - AC-001-1/B2: 目录含隐藏文件\n")
	addCHG(t, dir, "| CHG-002 | 2026-11-26 | 修改 | FR-001 | 补充用例 |")
	wantPass(t, checkJSON(t, dir))
}

func TestCheck_AC_019_3_E1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	edit(t, dir, "  - AC-001-1/E1: 只读目录\n", "")
	addCHG(t, dir, "| CHG-002 | 2026-11-26 | 修改 | FR-001 | 删用例 |")
	wantViolation(t, checkJSON(t, dir), "prd.03", "AC-001-1/E1 从文档中消失")
}

func TestCheck_AC_019_3_E2(t *testing.T) {
	dir := lockedRepo(t, prd(t, "  - AC-001-1/E1: 只读目录", "  - AC-001-1/E1: Removed：不再考虑只读目录\n  - AC-001-1/B1: 路径含空格"))
	edit(t, dir, "  - AC-001-1/E1: Removed：不再考虑只读目录", "  - AC-001-1/E1: 磁盘已满")
	addCHG(t, dir, "| CHG-002 | 2026-11-26 | 修改 | FR-001 | 复用用例编号 |")
	wantViolation(t, checkJSON(t, dir), "prd.03", "AC-001-1/E1 已 Removed")
}
