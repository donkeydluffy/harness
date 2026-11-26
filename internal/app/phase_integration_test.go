package app_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// phaseRepo 返回带 PRD、架构与锁文件的仓库；config 非空时写入 harness.json。
func phaseRepo(t *testing.T, config string) string {
	t.Helper()
	dir := lockedRepo(t, basePRD)
	if config != "" {
		if err := os.WriteFile(filepath.Join(dir, "harness.json"), []byte(config), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// setPhase 把锁文件中的当前阶段改为 name（模拟人已推进到该阶段）。
func setPhase(t *testing.T, dir, name string) {
	t.Helper()
	p := filepath.Join(dir, "harness.lock")
	b, _ := os.ReadFile(p)
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	m["phase"] = name
	b, _ = json.Marshal(m)
	os.WriteFile(p, b, 0o644)
}

func lockPhase(t *testing.T, dir string) string {
	t.Helper()
	var l struct{ Phase string }
	json.Unmarshal([]byte(readFile(t, dir, "harness.lock")), &l)
	return l.Phase
}

const releasePhase = `{"name": "release", "after": "implementation", "entry": [], "exit": ["prd.*"], "writable": ["docs/release/**"], "skill": "release"}`

// ---- AC-045-1 阶段定义 ----

func TestCheck_AC_045_1_N1(t *testing.T) {
	wantPass(t, checkJSON(t, phaseRepo(t, ""), "phase"))
	r := runHarness(t, phaseRepo(t, ""), "phase", "list")
	for _, name := range []string{"requirements", "architecture", "planning", "implementation"} {
		if !strings.Contains(r.Stdout, name) {
			t.Fatalf("phase list 缺少 %s：%s", name, r.Stdout)
		}
	}
}

func TestCheck_AC_045_1_B1(t *testing.T) {
	cfg := `{"phases": [{"name": "intake", "after": "requirements", "entry": [], "exit": ["prd.*"], "writable": ["docs/intake/**"], "skill": "intake"}]}`
	wantPass(t, checkJSON(t, phaseRepo(t, cfg), "phase")) // 进入规则为空列表
	r := runHarness(t, phaseRepo(t, ""), "phase", "list")
	if !strings.HasPrefix(r.Stdout, "* 1 requirements") {
		t.Fatalf("第一个阶段应为当前阶段：%s", r.Stdout)
	}
}

func TestCheck_AC_045_1_E1(t *testing.T) {
	cfg := `{"phases": [{"name": "release", "after": "implementation", "entry": [], "writable": ["docs/release/**"], "skill": "release"}]}`
	wantViolation(t, checkJSON(t, phaseRepo(t, cfg), "phase"), "phase.01", "release", "缺少 exit")
}

func TestCheck_AC_045_1_E2(t *testing.T) {
	cfg := strings.Replace(`{"phases": [`+releasePhase+`]}`, `"prd.*"`, `"zzz.09"`, 1)
	wantViolation(t, checkJSON(t, phaseRepo(t, cfg), "phase"), "phase.01", "zzz.09 不存在")
}

// ---- AC-045-2 项目配置扩展阶段 ----

func TestCheck_AC_045_2_N1(t *testing.T) {
	dir := phaseRepo(t, `{"phases": [`+releasePhase+`]}`)
	wantPass(t, checkJSON(t, dir, "phase"))
	if r := runHarness(t, dir, "phase", "list"); !strings.Contains(r.Stdout, "5 release") {
		t.Fatalf("phase list 缺少 release：%s", r.Stdout)
	}
}

func TestCheck_AC_045_2_N2(t *testing.T) {
	// Release 1 没有项目规则（FR-046 为 Release 2），用内置规则 debt.01 代替被追加的规则。
	dir := phaseRepo(t, `{"phases": [{"name": "architecture", "addExit": ["debt.01"]}]}`)
	os.WriteFile(filepath.Join(dir, "docs/debt.md"), []byte("# 技术债清单\n\n### DEBT-001 x\n- Status: Done\n- Module: MOD-01\n\n说明。\n"), 0o644)
	setPhase(t, dir, "architecture")
	wantPass(t, checkJSON(t, dir, "phase"))
	out, code := inTerminal(t, dir, "", "", "phase", "next")
	if code != 1 || !strings.Contains(out, "debt.01") {
		t.Fatalf("追加的退出规则应在推进时执行：%d %s", code, out)
	}
}

func TestCheck_AC_045_2_E1(t *testing.T) {
	dir := phaseRepo(t, `{"phases": [{"name": "architecture", "removeExit": ["adr.*"]}]}`)
	wantViolation(t, checkJSON(t, dir, "phase"), "phase.01", "architecture", "adr.*", "ADR")
	// 引用存在的 ADR 后通过。
	os.WriteFile(filepath.Join(dir, "harness.json"), []byte(`{"phases": [{"name": "architecture", "removeExit": ["adr.*"], "adr": "ADR-001"}]}`), 0o644)
	os.MkdirAll(filepath.Join(dir, "docs/adr"), 0o755)
	os.WriteFile(filepath.Join(dir, "docs/adr/ADR-001-x.md"), []byte(sprintf(adr, "Accepted", "FR-001", "")), 0o644)
	wantPass(t, checkJSON(t, dir, "phase"))
}

func TestCheck_AC_045_2_E2(t *testing.T) {
	cfg := strings.Replace(`{"phases": [`+releasePhase+`]}`, `"name": "release"`, `"name": "architecture"`, 1)
	wantViolation(t, checkJSON(t, phaseRepo(t, cfg), "phase"), "phase.01", "architecture", "同名")
}

// ---- AC-045-3 推进阶段（E3 门禁部分由 T-016 实现） ----

func TestPhase_AC_045_3_N1(t *testing.T) {
	dir := phaseRepo(t, "")
	out, code := inTerminal(t, dir, confirmPrompt, "y", "phase", "next")
	if code != 0 || lockPhase(t, dir) != "architecture" {
		t.Fatalf("应推进到 architecture：%d %s，锁文件阶段 %q", code, out, lockPhase(t, dir))
	}
}

func TestPhase_AC_045_3_E1(t *testing.T) {
	dir := phaseRepo(t, "")
	edit(t, dir, "- Type: Constraint\n", "")
	out, code := inTerminal(t, dir, "", "", "phase", "next")
	if code != 1 || !strings.Contains(out, "prd.02") || !strings.Contains(out, "拒绝推进") || lockPhase(t, dir) != "" {
		t.Fatalf("退出规则未通过应拒绝：%d %s", code, out)
	}
}

func TestPhase_AC_045_3_E2(t *testing.T) {
	dir := phaseRepo(t, "")
	r := runHarness(t, dir, "phase", "next")
	if r.Code == 0 || !strings.Contains(r.Stderr, "交互式终端") || lockPhase(t, dir) != "" {
		t.Fatalf("非交互环境应拒绝：%d %s", r.Code, r.Stderr)
	}
}

func TestPhase_AC_045_3_B1(t *testing.T) {
	dir := phaseRepo(t, "")
	setPhase(t, dir, "implementation")
	out, code := inTerminal(t, dir, "", "", "phase", "next")
	if code == 0 || !strings.Contains(out, "最后") {
		t.Fatalf("最后阶段应提示并以非 0 退出：%d %s", code, out)
	}
}
