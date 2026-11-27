package app_test

import (
	"os"
	"testing"
)

// extDemoRepo 建立使用本仓库真实架构总览的 git 仓库（EXT-03 的 Paths 与 Closed 取自 docs/architecture/overview.md），
// 任务 T-901 扩展 ext，提交 files；返回目录。
func extDemoRepo(t *testing.T, ext, scope string, files map[string]string) string {
	t.Helper()
	overview, err := os.ReadFile("../../docs/architecture/overview.md")
	if err != nil {
		t.Fatal(err)
	}
	tasks := "# 任务列表\n\n- Status: Draft\n\n### T-901 示例扩展\n- Status: InProgress\n- Scope: " + scope + "\n- Extends: " + ext + "\n"
	dir := gitRepo(t, map[string]string{
		"docs/architecture/overview.md":      string(overview),
		"docs/tasks.md":                      tasks,
		"internal/gate/gate.go":              "package gate\n",
		"internal/rules/engine/engine.go":    "package engine\n",
		"internal/rules/builtin/registry.go": "package builtin\n",
	})
	stage(t, dir, files)
	git(t, dir, "commit", "-q", "-m", "示例扩展\n\nTask: T-901")
	return dir
}

// 示例规则 demo.01：规则实现文件、规则样本目录、规则注册表文件（EXT-03）。
var demoRule = map[string]string{
	"internal/rules/builtin/demo_01.go":                      "package builtin\n\n// demo01 是示例规则。\ntype demo01 struct{}\n",
	"internal/rules/builtin/testdata/demo.01/ok/a.md":        "ok\n",
	"internal/rules/builtin/testdata/demo.01/violation/a.md": "bad\n",
	"internal/rules/builtin/registry.go":                     "package builtin\n\n// demo01{}\n",
}

const ruleScope = "internal/rules/builtin/**"

// ---- AC-N009-1 新增内置规则 ----

func TestCheck_AC_N009_1_N1(t *testing.T) {
	wantPass(t, checkJSON(t, extDemoRepo(t, "EXT-03", ruleScope, demoRule), "commits", "HEAD~1..HEAD", "commit.04"))
}

func TestCheck_AC_N009_1_E1(t *testing.T) {
	files := map[string]string{"internal/rules/engine/engine.go": "package engine\n\n// 为示例规则加特例\n"}
	for k, v := range demoRule {
		files[k] = v
	}
	c := checkJSON(t, extDemoRepo(t, "EXT-03", ruleScope, files), "commits", "HEAD~1..HEAD", "commit.04")
	wantViolation(t, c, "commit.04", "internal/rules/engine/engine.go")
}
