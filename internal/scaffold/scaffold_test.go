package scaffold_test

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/donkeydluffy/harness/internal/scaffold"
	"github.com/donkeydluffy/harness/internal/workspace"
)

func TestGenerate_IF_020(t *testing.T) {
	dir := t.TempDir()
	opt := scaffold.InitOptions{Languages: []string{"go"}, CI: "github", CIVersion: "0.2.0"}
	r, err := scaffold.Generate(dir, opt)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"harness.json", ".claude/settings.json", ".github/workflows/harness.yml", "docs/PRD.md"} {
		if !slices.Contains(r.Written, want) {
			t.Errorf("应写入 %s，实际：%v", want, r.Written)
		}
	}
	// 再次生成：内容相同的文件不动；改过的文件跳过且不被覆盖。
	edited := filepath.Join(dir, "docs/PRD.md")
	if err := os.WriteFile(edited, []byte("# 已修改\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err = scaffold.Generate(dir, opt)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Written) != 0 || !slices.Contains(r.Skipped, "docs/PRD.md") {
		t.Errorf("第二次生成：Written %v，Skipped %v", r.Written, r.Skipped)
	}
	if b, _ := os.ReadFile(edited); string(b) != "# 已修改\n" {
		t.Errorf("改过的文件被覆盖：%q", b)
	}
	if _, err := scaffold.Generate(t.TempDir(), scaffold.InitOptions{Languages: []string{"cobol"}}); err == nil {
		t.Error("不支持的语言应返回错误")
	}
}

func TestExpected_IF_021(t *testing.T) {
	dir := t.TempDir()
	if _, err := scaffold.Generate(dir, scaffold.InitOptions{Languages: []string{"go"}, CI: "github", CIVersion: "0.2.0"}); err != nil {
		t.Fatal(err)
	}
	s, err := workspace.Load(dir, workspace.LoadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want, err := scaffold.Expected(s)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{".claude/settings.json", ".github/workflows/harness.yml"} {
		got, err := os.ReadFile(filepath.Join(dir, p))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(want[p], got) {
			t.Errorf("%s 的期望内容与生成内容不一致", p)
		}
	}
}
