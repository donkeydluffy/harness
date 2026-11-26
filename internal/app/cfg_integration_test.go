package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"harness/internal/scaffold"
)

// addGenerated 补上 harness init 生成的 hook 配置与 CI 配置：有 harness.json 的项目视为已初始化，
// cfg.02 要求生成文件一致，ci.01 要求有 CI 配置。
func addGenerated(t *testing.T, files map[string]string) {
	t.Helper()
	gen, err := scaffold.Files(scaffold.InitOptions{Languages: []string{"go"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{".claude/settings.json", ".github/workflows/harness.yml"} {
		files[p] = string(gen[p])
	}
}

const cfgJSON = `{
  "commitMaxLines": 400,
  "vagueWords": ["快速", "友好", "适当"]
}
`

// cfgRepo 返回带 harness.json 并已执行 harness lock 的项目。
func cfgRepo(t *testing.T) string {
	t.Helper()
	files := traced(t, basePRD)
	files["harness.json"] = cfgJSON
	addGenerated(t, files)
	dir := writeTree(t, files)
	if r := runHarness(t, dir, "lock"); r.Code != 0 {
		t.Fatalf("lock：%s%s", r.Stdout, r.Stderr)
	}
	return dir
}

func setConfig(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "harness.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// ---- AC-041-1 门禁配置纳入锁文件 ----

func TestCheck_AC_041_1_N1(t *testing.T) {
	dir := cfgRepo(t)
	setConfig(t, dir, strings.Replace(cfgJSON, "400", "300", 1))
	addCHG(t, dir, "| CHG-002 | 2026-11-26 | 修改 | harness.json | 提交行数上限改为 300 |")
	wantPass(t, checkJSON(t, dir, "cfg.01", "prd"))
	// harness lock 之后，同一配置不再需要新的 CHG；再改一次又需要。
	if r := runHarness(t, dir, "lock"); r.Code != 0 {
		t.Fatal(r.Stderr)
	}
	setConfig(t, dir, strings.Replace(cfgJSON, "400", "200", 1))
	wantViolation(t, checkJSON(t, dir, "cfg.01"), "cfg.01", "harness.json", "没有引用它的新 CHG 条目或 ADR")
}

func TestCheck_AC_041_1_E1(t *testing.T) {
	dir := cfgRepo(t)
	setConfig(t, dir, strings.Replace(cfgJSON, `, "适当"`, "", 1))
	wantViolation(t, checkJSON(t, dir, "cfg.01"), "cfg.01", "内容有变化", "没有引用它的新 CHG 条目或 ADR")
}

func TestCheck_AC_041_1_B1(t *testing.T) {
	dir := cfgRepo(t)
	// 只加注释项、调整缩进与键顺序，规范化哈希不变。
	setConfig(t, dir, `{"//": "团队约定：上限不低于 300", "vagueWords": ["快速", "友好", "适当"],
	"commitMaxLines": 400}`)
	wantPass(t, checkJSON(t, dir, "cfg.01"))
}

// 新增的 ADR 在 Relates 中引用 harness.json 时，同样可以说明配置变化。
func TestCheck_ConfigChangeWithADR(t *testing.T) {
	dir := cfgRepo(t)
	setConfig(t, dir, strings.Replace(cfgJSON, "400", "300", 1))
	write(t, dir, "docs/adr/ADR-001.md", sprintf(adr, "Accepted", "FR-001, harness.json", ""))
	wantPass(t, checkJSON(t, dir, "cfg.01"))
}

// ---- AC-043-2 CI 配置包含 harness check ----

func TestCheck_AC_043_2_N1(t *testing.T) {
	wantPass(t, checkJSON(t, initDir(t, "go"), "ci.01"))
}

func TestCheck_AC_043_2_E1(t *testing.T) {
	dir := initDir(t, "go")
	write(t, dir, ".github/workflows/harness.yml", "name: build\non: [push]\njobs:\n  b:\n    runs-on: ubuntu-latest\n    steps:\n      - run: go install example.com/harness@latest && harness lock\n")
	write(t, dir, ".github/workflows/lint.yaml", "name: lint\n")
	wantViolation(t, checkJSON(t, dir, "ci.01"), "ci.01", "都不包含 harness check", ".github/workflows/harness.yml", ".github/workflows/lint.yaml")
}

func TestCheck_AC_043_2_B1(t *testing.T) {
	dir := initDir(t, "go")
	os.Remove(filepath.Join(dir, ".github/workflows/harness.yml"))
	wantViolation(t, checkJSON(t, dir, "ci.01"), "ci.01", "没有任何 CI 配置", "harness init")
}
