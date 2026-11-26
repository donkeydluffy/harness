package app_test

import (
	"strings"
	"testing"
)

const (
	ownerRule = `{"id": "acme.01", "kind": "required", "entity": "FR", "field": "Owner", "message": "写明负责人", "samples": "rules/acme.01"}`
	printRule = `{"id": "acme.02", "kind": "forbid", "paths": ["internal/**"], "pattern": "fmt\\.Println", "message": "用日志", "samples": "rules/acme.02"}`
	tbdRule   = `{"id": "acme.03", "kind": "forbid", "paths": ["docs/**"], "pattern": "待定", "message": "设计不留待定项", "samples": "rules/acme.03", "approveFor": ["design"]}`
	frOnly    = "# PRD\n\n- Status: Draft\n- Current-Release: 1\n\n### FR-001 a\n- Status: Active\n"
)

// ruleSamples 返回三条规则的违规样本与合规样本。
func ruleSamples() map[string]string {
	return map[string]string{
		"rules/acme.01/violation/docs/PRD.md":   frOnly,
		"rules/acme.01/ok/docs/PRD.md":          frOnly + "- Owner: alice\n",
		"rules/acme.02/violation/internal/a.go": "package a\n\nfunc f() { fmt.Println(1) }\n",
		"rules/acme.02/ok/internal/a.go":        "package a\n",
		"rules/acme.03/violation/docs/PRD.md":   frOnly + "\n范围待定。\n",
		"rules/acme.03/ok/docs/PRD.md":          frOnly,
	}
}

func rulesConfig(defs ...string) string {
	return "{\n  \"rules\": [" + strings.Join(defs, ", ") + "]\n}\n"
}

// customRepo 返回登记了 defs 的项目（PRD 为 basePRD，其中的 FR 没有 Owner 字段）。
func customRepo(t *testing.T, defs ...string) string {
	t.Helper()
	files := ruleSamples()
	files["docs/PRD.md"] = basePRD
	files["harness.json"] = rulesConfig(defs...)
	files["internal/x/x.go"] = "package x\n\nimport \"fmt\"\n\nfunc F() { fmt.Println(1) }\n"
	return writeTree(t, files)
}

// ---- AC-054-1 声明式规则 ----

func TestCheck_AC_054_1_N1(t *testing.T) {
	c := checkJSON(t, customRepo(t, ownerRule), "acme")
	v := wantViolation(t, c, "acme.01", "FR-001 缺少字段 Owner", "写明负责人")
	if v.File != "docs/PRD.md" {
		t.Fatalf("位置：%+v", v)
	}
	wantViolation(t, c, "acme.01", "FR-002 缺少字段 Owner")
}

func TestCheck_AC_054_1_N2(t *testing.T) {
	c := checkJSON(t, customRepo(t, printRule), "acme.02")
	if v := wantViolation(t, c, "acme.02", "fmt\\.Println"); v.File != "internal/x/x.go" || v.Line != 5 {
		t.Fatalf("应报告文件与行号：%+v", v)
	}
}

func TestCheck_AC_054_1_E1(t *testing.T) {
	c := checkJSON(t, customRepo(t, strings.Replace(ownerRule, `"required"`, `"magic"`, 1)), "acme")
	wantViolation(t, c, "acme.01", "不支持", "magic")
}

// ---- AC-054-2 登记并纳入锁文件 ----

func TestCheck_AC_054_2_N1(t *testing.T) {
	r := runHarness(t, customRepo(t, ownerRule), "check")
	if r.Code != 1 || !strings.Contains(r.Stdout, "[acme.01]") {
		t.Fatalf("harness check 应运行登记的规则：%d %s", r.Code, r.Stdout)
	}
}

func TestCheck_AC_054_2_E1(t *testing.T) {
	files := traced(t, basePRD)
	for k, v := range ruleSamples() {
		files[k] = v
	}
	files["harness.json"] = rulesConfig(printRule)
	addGenerated(t, files)
	dir := writeTree(t, files)
	lockOK(t, dir)
	wantPass(t, checkJSON(t, dir, "acme"))
	editFile(t, dir, "harness.json", "用日志", "改用日志")
	wantViolation(t, checkJSON(t, dir, "acme"), "acme.02", "与锁文件不一致")
	editFile(t, dir, "docs/PRD.md", "| CHG-001 | 2026-11-25 | 新增 | NFR-001, NFR-002 | 初稿 |", "| CHG-001 | 2026-11-25 | 新增 | NFR-001, NFR-002 | 初稿 |\n| CHG-002 | 2026-11-26 | 修改 | harness.json | 规则说明 |")
	lockOK(t, dir)
	editFile(t, dir, "rules/acme.02/ok/internal/a.go", "package a\n", "package a // ok\n")
	wantViolation(t, checkJSON(t, dir, "acme"), "acme.02", "与锁文件不一致")
}

// ---- AC-054-3 样本 ----

func TestCheck_AC_054_3_N1(t *testing.T) {
	c := checkJSON(t, customRepo(t, ownerRule), "acme")
	for _, v := range c.vs {
		if strings.Contains(v.Msg, "失效") {
			t.Fatalf("样本符合预期时规则不应失效：%+v", v)
		}
	}
	wantViolation(t, c, "acme.01", "缺少字段 Owner")
}

func TestCheck_AC_054_3_E1(t *testing.T) {
	dir := customRepo(t, ownerRule)
	editFile(t, dir, "rules/acme.01/violation/docs/PRD.md", "- Status: Active\n", "- Status: Active\n- Owner: bob\n")
	c := checkJSON(t, dir, "acme")
	wantViolation(t, c, "acme.01", "违规样本", "失效")
	for _, v := range c.vs {
		if strings.Contains(v.Msg, "缺少字段") {
			t.Fatalf("失效的规则不应运行：%+v", v)
		}
	}
}

func TestCheck_AC_054_3_E2(t *testing.T) {
	files := ruleSamples()
	delete(files, "rules/acme.01/ok/docs/PRD.md")
	files["docs/PRD.md"] = basePRD
	files["harness.json"] = rulesConfig(ownerRule)
	wantViolation(t, checkJSON(t, writeTree(t, files), "acme"), "acme.01", "缺少样本 rules/acme.01/ok")
}

// ---- AC-054-4 编号前缀 ----

func TestCheck_AC_054_4_N1(t *testing.T) {
	c := checkJSON(t, customRepo(t, ownerRule), "acme")
	for _, v := range c.vs {
		if strings.Contains(v.Msg, "定义有误") {
			t.Fatalf("acme.01 应通过编号校验：%+v", v)
		}
	}
}

func TestCheck_AC_054_4_E1(t *testing.T) {
	c := checkJSON(t, customRepo(t, strings.Replace(ownerRule, "acme.01", "prd.01", 1)), "prd.01")
	wantViolation(t, c, "prd.01", "与内置规则冲突")
}

func TestCheck_AC_054_4_E2(t *testing.T) {
	r := runHarness(t, customRepo(t, strings.Replace(ownerRule, `"acme.01"`, `"01"`, 1)), "check")
	if r.Code != 1 || !strings.Contains(r.Stdout, "编号须为 <项目前缀>.<两位以上序号>") {
		t.Fatalf("没有前缀的编号应失败：%d %s", r.Code, r.Stdout)
	}
}

// ---- AC-054-5 批准前检查 ----

// tbdRepo 返回登记了 acme.03（设计元素的批准前检查）的第 2 版锁文件仓库；FR-002 与 MOD-02 都含"待定"。
func tbdRepo(t *testing.T) string {
	t.Helper()
	dir := unitRepo(t)
	for k, v := range ruleSamples() {
		writeFile(t, dir, k, v)
	}
	writeFile(t, dir, "harness.json", rulesConfig(tbdRule))
	editFile(t, dir, "docs/PRD.md", "- AC-002-1: 记录问答。", "- AC-002-1: 记录问答，保留期限待定。",
		"| CHG-001 | 2026-11-25 | 新增 | NFR-001, NFR-002 | 初稿 |", "| CHG-001 | 2026-11-25 | 新增 | NFR-001, NFR-002 | 初稿 |\n| CHG-002 | 2026-11-26 | 修改 | FR-002 | 保留期限 |")
	editFile(t, dir, "docs/architecture/overview.md", "- Coverage: 70\n", "- Coverage: 70\n\n存储实现待定。\n")
	return dir
}

func TestApprove_AC_054_5_N1(t *testing.T) {
	dir := tbdRepo(t)
	mustApprove(t, dir, "prd")
	out, code := approveUnit(t, dir, "MOD-02", "y")
	if code != 1 || !strings.Contains(out, "acme.03") || !strings.Contains(out, "拒绝批准") {
		t.Fatalf("违反批准前检查的模块应被拒绝：%d %s", code, out)
	}
	if unitRecords(t, dir)["MOD-02"] != nil {
		t.Fatal("MOD-02 不应被批准")
	}
	mustApprove(t, dir, "MOD-01") // 违规只落在 MOD-02 内，不影响其他设计元素
}

func TestApprove_AC_054_5_N2(t *testing.T) {
	dir := tbdRepo(t)
	mustApprove(t, dir, "prd")
	if unitRecords(t, dir)["FR-002"] == nil {
		t.Fatal("登记在设计元素上的规则不应影响需求的批准")
	}
}

func TestCheck_AC_054_5_E1(t *testing.T) {
	r := runHarness(t, customRepo(t, strings.Replace(tbdRule, `["design"]`, `["module"]`, 1)), "check", "acme")
	if r.Code != 1 || !strings.Contains(r.Stdout, `批准单位类型 "module" 不存在`) {
		t.Fatalf("不存在的批准单位类型应失败：%d %s", r.Code, r.Stdout)
	}
}
