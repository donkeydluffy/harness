package app_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	goMod     = "module demo\n\ngo 1.21\n"
	goClean   = "package a\n\n// A 是示例。\nfunc A() int { return 1 }\n"
	goUnfmt   = "package a\n\nfunc   B()   int { return 2 }\n"
	goVetBad  = "package a\n\nimport \"fmt\"\n\nfunc C() { fmt.Printf(\"%d\\n\", \"x\") }\n"
	goLintCfg = "{\n  \"languages\": [\"go\"],\n  \"sourcePaths\": [\"internal/**\"]\n}\n"
)

// lintRepo 返回 Go 项目；extra 中的文件追加写入。
func lintRepo(t *testing.T, cfg string, extra map[string]string) string {
	t.Helper()
	files := map[string]string{"go.mod": goMod, "harness.json": cfg, "internal/a/a.go": goClean}
	for k, v := range extra {
		files[k] = v
	}
	return writeTree(t, files)
}

// runWithPath 以指定的 PATH 运行 harness；path 为空时沿用当前环境。
func runWithPath(t *testing.T, dir, path string, args ...string) result {
	t.Helper()
	cmd := exec.Command(harnessBin, args...)
	cmd.Dir = dir
	if path != "" {
		cmd.Env = append(os.Environ(), "PATH="+path)
	}
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return result{out.String(), errb.String(), code}
}

type lintIssue struct {
	File, Tool, Msg string
	Line            int
}

func lintIssues(t *testing.T, dir string) []lintIssue {
	t.Helper()
	var r struct {
		SourceHash string
		Issues     []lintIssue
	}
	if err := json.Unmarshal([]byte(readFile(t, dir, ".harness/reports/lint.json")), &r); err != nil || r.SourceHash == "" {
		t.Fatalf("lint 报告：%v %+v", err, r)
	}
	return r.Issues
}

func hasIssue(issues []lintIssue, file, tool string) *lintIssue {
	for i, is := range issues {
		if is.File == file && is.Tool == tool {
			return &issues[i]
		}
	}
	return nil
}

// noTools 是不含 ruff、flake8 等工具的 PATH。
const noTools = "/usr/bin:/bin"

// ---- AC-014-2 工具缺失 ----

func TestLint_AC_014_2_N1(t *testing.T) {
	dir := lintRepo(t, goLintCfg, nil)
	if r := runHarness(t, dir, "check", "lint"); r.Code != 0 {
		t.Fatalf("工具齐全且代码合规时应通过：%d %s%s", r.Code, r.Stdout, r.Stderr)
	}
	if len(lintIssues(t, dir)) != 0 {
		t.Fatal("报告不应有问题")
	}
}

func TestLint_AC_014_2_E1(t *testing.T) {
	dir := writeTree(t, map[string]string{"harness.json": `{"languages": ["python"], "sourcePaths": ["src/**"]}`, "src/a.py": "x = 1\n"})
	r := runWithPath(t, dir, noTools, "check", "lint")
	if r.Code != 1 || !strings.Contains(r.Stderr, "ruff") || !strings.Contains(r.Stderr, "pip install ruff") {
		t.Fatalf("未安装 ruff 时应失败并给出安装命令：%d %s", r.Code, r.Stderr)
	}
}

func TestLint_AC_014_2_B1(t *testing.T) {
	cfg := `{"languages": ["python"], "sourcePaths": ["src/**"], "tools": {"python": {"format": "true", "lint": "flake8 src"}}}`
	r := runWithPath(t, writeTree(t, map[string]string{"harness.json": cfg, "src/a.py": "x = 1\n"}), noTools, "check", "lint")
	if r.Code == 0 || !strings.Contains(r.Stderr, "flake8") || strings.Contains(r.Stderr, "ruff") {
		t.Fatalf("应提示安装 flake8 而不是 ruff：%d %s", r.Code, r.Stderr)
	}
}

// ---- AC-014-3 按语言覆盖命令 ----

func TestLint_AC_014_3_N1(t *testing.T) {
	dir := lintRepo(t, goLintCfg, map[string]string{"internal/a/b.go": goUnfmt, "internal/a/c.go": goVetBad})
	runHarness(t, dir, "check", "lint")
	issues := lintIssues(t, dir)
	if hasIssue(issues, "internal/a/b.go", "gofmt") == nil || hasIssue(issues, "internal/a/c.go", "go vet") == nil {
		t.Fatalf("未覆盖时应运行 gofmt -l 与 go vet：%+v", issues)
	}
}

func TestLint_AC_014_3_N2(t *testing.T) {
	bin := t.TempDir()
	os.WriteFile(filepath.Join(bin, "staticcheck"), []byte("#!/bin/sh\necho 'internal/a/a.go:4:1: fake finding'\nexit 1\n"), 0o755)
	cfg := `{"languages": ["go"], "sourcePaths": ["internal/**"], "tools": {"go": {"lint": "staticcheck ./..."}}}`
	dir := lintRepo(t, cfg, map[string]string{"internal/a/c.go": goVetBad})
	r := runWithPath(t, dir, bin+string(os.PathListSeparator)+os.Getenv("PATH"), "check", "lint")
	issues := lintIssues(t, dir)
	if is := hasIssue(issues, "internal/a/a.go", "staticcheck"); is == nil || is.Line != 4 || r.Code == 0 {
		t.Fatalf("应运行 staticcheck：%d %+v", r.Code, issues)
	}
	for _, is := range issues {
		if is.Tool == "go vet" {
			t.Fatalf("覆盖后不应运行 go vet：%+v", issues)
		}
	}
}

func TestLint_AC_014_3_E1(t *testing.T) {
	dir := lintRepo(t, `{"languages": ["go"], "tools": {"rust": {"lint": "cargo clippy"}}}`, nil)
	wantViolation(t, checkJSON(t, dir, "cfg.03"), "cfg.03", "rust")
}

// ---- AC-014-4 lint 报告 ----

func TestLint_AC_014_4_N1(t *testing.T) {
	dir := lintRepo(t, goLintCfg, map[string]string{"internal/a/b.go": goUnfmt})
	r := runHarness(t, dir, "check", "lint")
	if r.Code == 0 || hasIssue(lintIssues(t, dir), "internal/a/b.go", "gofmt") == nil {
		t.Fatalf("未格式化的文件应列入报告并失败：%d %s", r.Code, r.Stdout)
	}
}

func TestLint_AC_014_4_N2(t *testing.T) {
	dir := lintRepo(t, goLintCfg, map[string]string{"internal/a/c.go": goVetBad})
	runHarness(t, dir, "check", "lint")
	if is := hasIssue(lintIssues(t, dir), "internal/a/c.go", "go vet"); is == nil || is.Line != 5 || is.Msg == "" {
		t.Fatalf("go vet 的问题应带文件与行号：%+v", is)
	}
}

func TestLint_AC_014_4_B1(t *testing.T) {
	cfg := `{"languages": ["go", "python"], "sourcePaths": ["internal/**", "src/**"]}`
	dir := lintRepo(t, cfg, nil)
	path := filepath.Dir(goBin(t)) + string(os.PathListSeparator) + noTools
	if r := runWithPath(t, dir, path, "check", "lint"); r.Code != 0 {
		t.Fatalf("没有 Python 源文件时不应运行 Python 工具：%d %s%s", r.Code, r.Stdout, r.Stderr)
	}
}

func goBin(t *testing.T) string {
	p, err := exec.LookPath("go")
	if err != nil {
		t.Skip("没有 go")
	}
	return p
}

// ---- AC-014-5 全量检查读取 lint 报告 ----

func TestLint_AC_014_5_N1(t *testing.T) {
	dir := lintRepo(t, goLintCfg, nil)
	runHarness(t, dir, "check", "lint")
	wantPass(t, checkJSON(t, dir, "lint.01"))
}

func TestLint_AC_014_5_E1(t *testing.T) {
	dir := lintRepo(t, goLintCfg, map[string]string{"internal/a/c.go": goVetBad})
	runHarness(t, dir, "check", "lint")
	if v := wantViolation(t, checkJSON(t, dir, "lint.01"), "lint.01", "go vet"); v.File != "internal/a/c.go" || v.Line != 5 {
		t.Fatalf("应给出文件与行号：%+v", v)
	}
}

func TestLint_AC_014_5_E2(t *testing.T) {
	wantViolation(t, checkJSON(t, lintRepo(t, goLintCfg, nil), "lint.01"), "lint.01", "没有 lint 报告", "harness check lint")
}

func TestLint_AC_014_5_B1(t *testing.T) {
	dir := lintRepo(t, goLintCfg, nil)
	runHarness(t, dir, "check", "lint")
	editFile(t, dir, "internal/a/a.go", "return 1", "return 3")
	wantViolation(t, checkJSON(t, dir, "lint.01"), "lint.01", "过期")
}
