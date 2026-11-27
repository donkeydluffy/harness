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
