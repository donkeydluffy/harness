package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/donkeydluffy/harness/internal/lock"
	"github.com/donkeydluffy/harness/internal/scaffold"
)

// tree 在临时目录写入文件并返回目录。
func tree(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, c := range files {
		p := filepath.Join(dir, rel)
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

const badPRD = "# PRD\n\n## 功能需求\n\n### FR-001 示例\n- Status: Active\n- Release: 1\n- Priority: P0\n\n没有验收标准。\n"

func TestCheck_IF_022(t *testing.T) {
	dir := tree(t, map[string]string{"docs/PRD.md": badPRD})
	rep, err := Check(dir, CheckOptions{Rules: []string{"prd"}})
	if err != nil || len(rep.Violations) == 0 {
		t.Fatalf("需求没有验收标准时应有违规：%v %+v", err, rep)
	}
	if _, err := Check(dir, CheckOptions{Rules: []string{"nosuch.99"}}); err == nil || !strings.Contains(err.Error(), "nosuch.99") {
		t.Fatalf("选择器无效时应返回错误：%v", err)
	}
}

func TestLock_IF_023(t *testing.T) {
	dir := tree(t, map[string]string{"docs/PRD.md": badPRD})
	var ve *ViolationsError
	if err := Lock(dir); !errors.As(err, &ve) {
		t.Fatalf("有违规时拒绝锁定：%v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, lock.File)); !os.IsNotExist(err) {
		t.Fatal("拒绝锁定时不应写锁文件")
	}
	dir = t.TempDir()
	if _, err := Init(dir, scaffold.InitOptions{Languages: []string{"go"}, Version: "0.2.0"}); err != nil {
		t.Fatal(err)
	}
	if err := Lock(dir); err != nil {
		t.Fatalf("init 生成的骨架可以锁定：%v", err)
	}
	if l, err := lock.Read(filepath.Join(dir, lock.File)); err != nil || len(l.Items) == 0 {
		t.Fatalf("应写入锁文件：%v %+v", err, l)
	}
}

type noTTY struct{}

func (noTTY) IsInteractive() bool { return false }
func (noTTY) Confirm(string) bool { return true }
func (noTTY) Identity() string    { return "t" }

func TestApprove_IF_024(t *testing.T) {
	if err := Approve(t.TempDir(), "prd", noTTY{}); !errors.Is(err, ErrNotInteractive) {
		t.Fatalf("不在交互式终端中时拒绝批准：%v", err)
	}
}

func TestInit_IF_026(t *testing.T) {
	dir := t.TempDir()
	res, err := Init(dir, scaffold.InitOptions{Languages: []string{"go"}, Version: "0.2.0"})
	if err != nil || len(res.Written) == 0 {
		t.Fatalf("init：%v %+v", err, res)
	}
	if _, err := os.Stat(filepath.Join(dir, "harness.json")); err != nil {
		t.Fatal("应生成 harness.json")
	}
	if _, err := Init(t.TempDir(), scaffold.InitOptions{Languages: []string{"cobol"}, Version: "0.2.0"}); err == nil {
		t.Fatal("不支持的语言应返回错误")
	}
}

func TestSnapshot_IF_039(t *testing.T) {
	dir := tree(t, map[string]string{
		"harness.json": `{"languages": ["go"], "sourcePaths": ["internal/**"]}`, "go.mod": "module demo\n\ngo 1.21\n",
		"internal/a/a.go": "package a\n\n// harness:IF-001\nfunc A(x int) error { return nil }\n",
	})
	if err := Snapshot(dir); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, InterfacesFile))
	if string(b) != "IF-001\nfunc A(x int) error\n" {
		t.Fatalf("快照：%q", b)
	}
}

type yesTTY struct{}

func (yesTTY) IsInteractive() bool { return true }
func (yesTTY) Confirm(string) bool { return true }
func (yesTTY) Identity() string    { return "t" }

const approveOverview = "# 架构总览\n\n- Status: Draft\n\n### MOD-01 a\n- Status: Active\n- Paths: internal/a/**\n- Public: internal/a/a.go\n- Depends:\n- Interfaces: IF-001\n- Covers: FR-001\n- Core: false\n- Coverage: 0\n"

// approvedRepo 返回 init 生成并锁定的 Go 项目：MOD-01 的接口 IF-001 签名为 sig，代码为 code；PRD 与设计都已批准。
func approvedRepo(t *testing.T, sig, code string, tasks string) string {
	t.Helper()
	dir := t.TempDir()
	if _, err := Init(dir, scaffold.InitOptions{Languages: []string{"go"}, Version: "0.2.0"}); err != nil {
		t.Fatal(err)
	}
	if err := Lock(dir); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"go.mod":                                "module demo\n\ngo 1.21\n",
		"docs/architecture/overview.md":         approveOverview,
		"docs/architecture/modules/MOD-01-a.md": "# MOD-01 a\n\n### IF-001 A\n- Module: MOD-01\n- Covers: FR-001\n\n```go\n// IF-001\n" + sig + "\n```\n",
		"internal/a/a.go":                       "package a\n\n// harness:IF-001\n" + code + " { return nil }\n",
	}
	if tasks != "" {
		files["docs/tasks.md"] = "# 任务列表\n\n- Status: Draft\n\n" + tasks
	}
	for k, v := range files {
		os.MkdirAll(filepath.Dir(filepath.Join(dir, k)), 0o755)
		if err := os.WriteFile(filepath.Join(dir, k), []byte(v), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := Snapshot(dir); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"prd", "docs/architecture/overview.md", "docs/architecture/modules/MOD-01-a.md"} {
		if err := Approve(dir, target, yesTTY{}); err != nil {
			t.Fatalf("批准 %s：%v", target, err)
		}
	}
	return dir
}

func replaceIn(t *testing.T, dir, rel, old, new string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), old) {
		t.Fatalf("%s 中没有 %q", rel, old)
	}
	os.WriteFile(p, []byte(strings.Replace(string(b), old, new, 1)), 0o644)
}

// 设计签名先于代码修改：批准不比对代码（if.03 属于下游）。
func TestApprove_IF_024_SignatureBeforeCode(t *testing.T) {
	dir := approvedRepo(t, "func A() error", "func A() error", "")
	replaceIn(t, dir, "docs/architecture/modules/MOD-01-a.md", "func A() error", "func A(x int) error")
	if err := Approve(dir, "IF-001", yesTTY{}); err != nil {
		t.Fatalf("签名先改的接口应能批准：%v", err)
	}
}

// InProgress 任务的批准因上游变化失效：可直接重新批准（task.07 不参与自身的批准）。
func TestApprove_IF_024_InProgressTask(t *testing.T) {
	task := "### T-001 a\n- Status: InProgress\n- Covers: FR-001\n- Design: MOD-01\n- Scope: internal/a/**\n- Depends:\n"
	dir := approvedRepo(t, "func A() error", "func A() error", task)
	if err := Approve(dir, "T-001", yesTTY{}); err != nil {
		t.Fatalf("批准任务：%v", err)
	}
	replaceIn(t, dir, "docs/architecture/overview.md", "- Coverage: 0\n", "- Coverage: 10\n")
	if err := Approve(dir, "MOD-01", yesTTY{}); err != nil {
		t.Fatalf("批准 MOD-01：%v", err)
	}
	if err := Approve(dir, "T-001", yesTTY{}); err != nil {
		t.Fatalf("InProgress 任务应能直接重新批准：%v", err)
	}
}
