package gate

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestWriteTargets_IF_017(t *testing.T) {
	cases := []struct {
		args []string
		want []string
	}{
		{nil, nil},
		{[]string{"tee", "-a", "a.txt", "b.txt"}, []string{"a.txt", "b.txt"}},
		{[]string{"/bin/rm", "-f", "x"}, []string{"x"}},
		{[]string{"cp", "-r", "src", "dst"}, []string{"dst"}},
		{[]string{"cp"}, nil},
		{[]string{"mv", "a", "b"}, []string{"a", "b"}},
		{[]string{"sed", "-i", "", "s/a/b/", "f.go"}, []string{"f.go"}},
		{[]string{"sed", "-e", "s/a/b/", "--in-place=.bak", "g.go"}, []string{"g.go"}},
		{[]string{"sed", "-i", "--expression", "s/a/b/", "-f", "x.sed", "h.go"}, []string{"h.go"}},
		{[]string{"sed", "s/a/b/", "f.go"}, nil}, // 没有 -i 时只输出到标准输出
		{[]string{"git", "checkout", "main", "--", "a.go"}, []string{"a.go"}},
		{[]string{"git", "restore", "--staged", "b.go"}, []string{"b.go"}},
		{[]string{"git", "checkout", "main"}, nil},
		{[]string{"ls", "-l"}, nil},
	}
	for _, c := range cases {
		if got := writeTargets(c.args); !reflect.DeepEqual(got, c.want) {
			t.Errorf("writeTargets(%q) = %q，期望 %q", c.args, got, c.want)
		}
	}
}

func TestRegenRemovals_IF_017(t *testing.T) {
	s := load(t, project(t, map[string]string{".github/workflows/harness.yml": "x\n"}))
	if got := regenRemovals(s, "rm -f .github/workflows/harness.yml"); !got[".github/workflows/harness.yml"] || len(got) != 1 {
		t.Fatalf("删除生成文件应被识别：%v", got)
	}
	for _, cmd := range []string{
		"rm -r .github/workflows/harness.yml", // 其他选项
		"rm docs/PRD.md",                      // 不是生成文件
		"rm .github/workflows/*.yml",          // 通配符
		"rm -f",                               // 没有路径
		"rm a; rm b",                          // 多条命令
		"X=1 rm .gitlab-ci.yml",               // 带赋值
		"cat .gitlab-ci.yml",                  // 不是 rm
		"rm .gitlab-ci.yml > out",             // 带重定向
		"rm 'unterminated",                    // 解析失败
	} {
		if got := regenRemovals(s, cmd); got != nil {
			t.Errorf("%s 不应被识别为删除生成文件：%v", cmd, got)
		}
	}
}

func TestSnapshotRestore_IF_018(t *testing.T) {
	big := strings.Repeat("x", maxCopy+1)
	dir := project(t, map[string]string{"keep.txt": "原内容\n", "gone.txt": "将被删除\n", "big.bin": big})
	if changes, err := After(dir, "s1", func(string) bool { return false }); err != nil || changes != nil {
		t.Fatalf("没有命令前快照时应返回空：%v %v", changes, err)
	}
	if err := Before(dir, "s1"); err != nil {
		t.Fatal(err)
	}
	os.Remove(filepath.Join(dir, "gone.txt"))
	os.WriteFile(filepath.Join(dir, "big.bin"), []byte(big+"y"), 0o644)
	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("新文件\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("改动\n"), 0o644)
	changes, err := After(dir, "s1", func(p string) bool { return p == "keep.txt" })
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]Change{}
	for _, c := range changes {
		got[c.Path] = c
	}
	if c := got["gone.txt"]; c.Action != "deleted" || !c.Restored {
		t.Errorf("被删除的文件应恢复：%+v", c)
	}
	if c := got["new.txt"]; c.Action != "created" || !c.Restored {
		t.Errorf("新建的文件应删除：%+v", c)
	}
	if c := got["big.bin"]; c.Restored || !strings.Contains(c.Reason, "1 MiB") {
		t.Errorf("超过 1 MiB 的文件无法恢复：%+v", c)
	}
	if c := got["keep.txt"]; c.Reason != "允许" {
		t.Errorf("允许的改动应保留：%+v", c)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "gone.txt")); string(b) != "将被删除\n" {
		t.Errorf("恢复内容：%q", b)
	}
}
