package app_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// docRepo 返回带 basePRD 与一个源文件的项目；extra 中的文件覆盖或追加。
func docRepo(t *testing.T, extra map[string]string) string {
	t.Helper()
	files := map[string]string{
		"docs/PRD.md":               basePRD,
		"internal/api/a.go":         "package api\n",
		".claude/skills/x/SKILL.md": "# x\n\n结束前运行 `harness check`。\n",
	}
	for k, v := range extra {
		files[k] = v
	}
	return writeTree(t, files)
}

func checkDocs(t *testing.T, dir string) checked {
	t.Helper()
	return checkJSON(t, dir, "doc.02", "doc.03", "doc.04")
}

// ---- AC-030-1 行内代码中的路径 ----

func TestDoc_AC_030_1_N1(t *testing.T) {
	wantPass(t, checkDocs(t, docRepo(t, map[string]string{"README.md": "# demo\n\n需求见 `docs/PRD.md`。\n"})))
}

func TestDoc_AC_030_1_E1(t *testing.T) {
	dir := docRepo(t, map[string]string{".claude/skills/x/SKILL.md": "# x\n\n入口：\n\n见 `internal/old/x.go`。\n"})
	v := wantViolation(t, checkDocs(t, dir), "doc.02", "internal/old/x.go", "不存在")
	if v.File != ".claude/skills/x/SKILL.md" || v.Line != 5 {
		t.Fatalf("应给出文件与行号：%+v", v)
	}
}

func TestDoc_AC_030_1_B1(t *testing.T) {
	wantPass(t, checkDocs(t, docRepo(t, map[string]string{"README.md": "# demo\n\n```\ninternal/old/x.go\n`internal/old/y.go`\n```\n"})))
}

func TestDoc_AC_030_1_B2(t *testing.T) {
	wantPass(t, checkDocs(t, docRepo(t, map[string]string{"README.md": "# demo\n\n每个领域一篇 `docs/<领域>.md`。\n"})))
}

func TestDoc_AC_030_1_B3(t *testing.T) {
	prd := strings.Replace(basePRD, "  - AC-001-1/N1:", "  - AC-001-1/E9: `internal/old/x.go` 已不存在、引用 IF-099，失败\n  - AC-001-1/N1:", 1)
	tasks := "# 任务列表\n\n- Status: Draft\n\n### T-001 某任务\n- Status: Done\n- Covers: AC-001-1/E8\n- Design:\n- Scope: internal/api/**\n- Depends:\n"
	wantPass(t, checkDocs(t, docRepo(t, map[string]string{"docs/PRD.md": prd, "docs/tasks.md": tasks})))
}

func TestDoc_AC_030_1_B4(t *testing.T) {
	dir := t.TempDir()
	if r := runHarness(t, dir, "init", "--lang", "go", "--ci-version", "0.2.0"); r.Code != 0 {
		t.Fatalf("init 失败：%s%s", r.Stdout, r.Stderr)
	}
	wantPass(t, checkDocs(t, dir))
}

// ---- AC-030-2 harness 子命令 ----

func TestDoc_AC_030_2_N1(t *testing.T) {
	wantPass(t, checkDocs(t, docRepo(t, nil)))
}

func TestDoc_AC_030_2_E1(t *testing.T) {
	dir := docRepo(t, map[string]string{".claude/skills/x/SKILL.md": "# x\n\n先运行 `harness hook claudecode Stop`。\n"})
	wantViolation(t, checkDocs(t, dir), "doc.03", "hook", "不存在")
}

func TestDoc_AC_030_2_B1(t *testing.T) {
	prd := strings.Replace(basePRD, "## 4. 非功能需求", "### FR-009 度量\n- Status: Active\n- Release: 3\n- Priority: P2\n\n用 `harness stats` 输出度量。\n\n- AC-009-1: 输出度量。\n  - AC-009-1/N1: 输出\n\n## 4. 非功能需求", 1)
	wantPass(t, checkDocs(t, docRepo(t, map[string]string{"docs/PRD.md": prd})))
}

// ---- AC-030-3 编号引用 ----

func TestDoc_AC_030_3_N1(t *testing.T) {
	wantPass(t, checkDocs(t, docRepo(t, map[string]string{"docs/architecture/notes.md": "# 说明\n\n见 FR-002。\n"})))
}

func TestDoc_AC_030_3_E1(t *testing.T) {
	wantViolation(t, checkDocs(t, docRepo(t, map[string]string{"README.md": "# demo\n\n入口是 IF-099。\n"})), "doc.04", "IF-099", "不存在")
}

func TestDoc_AC_030_3_E2(t *testing.T) {
	prd := strings.Replace(basePRD, "- Current-Release: 1", "- Current-Release: 2", 1)
	prd = strings.Replace(prd, "### FR-002 澄清\n- Status: Active\n- Release: 1\n", "### FR-002 澄清\n- Status: Active\n- Release: 1\n- Replaced-By: FR-045\n", 1)
	prd = strings.Replace(prd, "## 4. 非功能需求", "### FR-045 新澄清\n- Status: Active\n- Release: 2\n- Priority: P1\n- Replaces: FR-002\n\n- AC-045-1: 澄清。\n  - AC-045-1/N1: 澄清\n\n## 4. 非功能需求", 1)
	dir := docRepo(t, map[string]string{"docs/PRD.md": prd, ".claude/skills/x/SKILL.md": "# x\n\n按 FR-002 澄清。\n"})
	wantViolation(t, checkDocs(t, dir), "doc.04", "FR-002", "已被 FR-045 取代")
}

func TestDoc_AC_030_3_B1(t *testing.T) {
	prd := strings.Replace(basePRD, "## 4. 非功能需求", "### FR-007 旧需求\n- Status: Removed\n- Release: 1\n- Priority: P2\n\n- AC-007-1: 旧。\n  - AC-007-1/N1: 旧\n\n## 4. 非功能需求", 1)
	prd = strings.TrimRight(prd, "\n") + "\n\n## 8. 变更日志\n\n| ID | 日期 | 动作 | 对象 | 说明 |\n|---|---|---|---|---|\n| CHG-009 | 2026-11-27 | 删除 | FR-007 | 不再需要 |\n"
	wantPass(t, checkDocs(t, docRepo(t, map[string]string{"docs/PRD.md": prd})))
}

// ---- AC-030-4 详细设计可能过期 ----

// staleRepo 提交模块 MOD-01 的详细设计后，再提交 n 次模块代码改动；返回 doc.05 的告警。
func staleRepo(t *testing.T, n int) string {
	t.Helper()
	dir := gitRepo(t, map[string]string{
		"harness.json":                          `{"languages": ["go"], "sourcePaths": ["internal/**"]}`,
		"docs/architecture/overview.md":         "# 架构总览\n\n- Status: Draft\n\n### MOD-01 a\n- Status: Active\n- Paths: internal/a/**\n- Public: internal/a/a.go\n- Depends:\n- Interfaces:\n- Covers: FR-001\n- Core: false\n- Coverage: 0\n",
		"docs/architecture/modules/MOD-01-a.md": "# MOD-01 a 详细设计\n\n- Status: Draft\n",
		"internal/a/a.go":                       "package a\n",
	})
	for i := 0; i < n; i++ {
		if err := os.WriteFile(filepath.Join(dir, "internal/a/a.go"), []byte(fmt.Sprintf("package a\n\nvar V = %d\n", i)), 0o644); err != nil {
			t.Fatal(err)
		}
		git(t, dir, "commit", "-q", "-am", fmt.Sprintf("改动 %d", i))
	}
	r := runHarness(t, dir, "check", "doc.05", "--json")
	if r.Code != 0 {
		t.Fatalf("告警不应影响退出码：%d %s%s", r.Code, r.Stdout, r.Stderr)
	}
	var rep struct{ Warnings []violation }
	if err := json.Unmarshal([]byte(r.Stdout), &rep); err != nil {
		t.Fatalf("解析输出：%v\n%s", err, r.Stdout)
	}
	var b strings.Builder
	for _, w := range rep.Warnings {
		b.WriteString(w.File + ": " + w.Msg + "\n")
	}
	return b.String()
}

func TestDoc_AC_030_4_N1(t *testing.T) {
	if w := staleRepo(t, 11); !strings.Contains(w, "docs/architecture/modules/MOD-01-a.md: MOD-01") || !strings.Contains(w, "11 个提交") {
		t.Fatalf("应告警设计可能过期：%q", w)
	}
}

func TestDoc_AC_030_4_B1(t *testing.T) {
	if w := staleRepo(t, 10); w != "" {
		t.Fatalf("恰好 10 次不应告警：%q", w)
	}
}
