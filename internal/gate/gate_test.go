package gate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/donkeydluffy/harness/internal/workspace"
)

func project(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, c := range files {
		p := filepath.Join(dir, rel)
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte(c), 0o644)
	}
	return dir
}

func load(t *testing.T, dir string) *workspace.Snapshot {
	t.Helper()
	s, err := workspace.Load(dir, workspace.LoadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

const unapprovedTask = "# 任务\n\n- Status: Draft\n\n### T-001 a\n- Status: InProgress\n- Covers: FR-001\n- Design: MOD-01\n- Scope: src/**\n- Depends:\n"

func TestDecide_IF_017(t *testing.T) {
	s := load(t, project(t, map[string]string{"harness.lock": `{"version": 2}`, "docs/tasks.md": unapprovedTask}))
	cases := []struct {
		path, rule string
	}{
		{"docs/notes.md", ""},             // 第 2 版锁文件下 docs/ 随时可写
		{"harness.lock", "gate.protect"},  // 保护文件
		{"src/a.go", "gate.write"},        // 任务批准链无效
		{"other/b.go", "gate.write"},      // 不在任务范围内
		{"../outside.go", ""},             // 仓库外
		{".harness/reports/go.jsonl", ""}, // 运行产物
	}
	for _, c := range cases {
		d := Decide(s, Request{Kind: Write, Paths: []string{c.path}})
		if d.Allow != (c.rule == "") || d.Rule != c.rule {
			t.Errorf("%s：%+v，期望规则 %q", c.path, d, c.rule)
		}
	}
	if d := Decide(s, Request{Kind: Command, Command: "harness approve prd"}); d.Allow {
		t.Error("代理执行 harness approve 应被拒绝")
	}
}

func TestSnapshot_IF_018(t *testing.T) {
	dir := project(t, map[string]string{"a.txt": "before\n", "keep.txt": "x\n"})
	if err := Before(dir, "s1"); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("after\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("y\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("n\n"), 0o644)
	changes, err := After(dir, "s1", func(p string) bool { return p == "keep.txt" })
	if err != nil || len(changes) != 3 {
		t.Fatalf("应有 3 处变化：%v %+v", err, changes)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "a.txt")); string(b) != "before\n" {
		t.Errorf("不允许的修改应恢复：%q", b)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "keep.txt")); string(b) != "y\n" {
		t.Errorf("允许的修改应保留：%q", b)
	}
	if _, err := os.Stat(filepath.Join(dir, "new.txt")); !os.IsNotExist(err) {
		t.Error("不允许的新建文件应删除")
	}
	if c, err := After(dir, "none", func(string) bool { return false }); err != nil || c != nil {
		t.Errorf("没有命令前快照时返回 nil：%v %v", c, err)
	}
}
