package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadConfig_IF_010(t *testing.T) {
	dir := t.TempDir()
	cfg, err := LoadConfig(filepath.Join(dir, "missing.json"))
	if err != nil || cfg.CommitMaxLines != 400 || len(cfg.VagueWords) != 12 {
		t.Fatalf("default: %+v %v", cfg, err)
	}
	write(t, dir, "harness.json", `{"commitMaxLines": 200, "vagueWords": ["快速"]}`)
	cfg, err = LoadConfig(filepath.Join(dir, "harness.json"))
	if err != nil || cfg.CommitMaxLines != 200 || len(cfg.VagueWords) != 1 || cfg.StopMaxBlocks != 3 {
		t.Fatalf("override: %+v %v", cfg, err)
	}
	write(t, dir, "bad.json", `{`)
	if _, err := LoadConfig(filepath.Join(dir, "bad.json")); err == nil {
		t.Fatal("want parse error")
	}
}

func TestLoad_IF_009(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "docs/PRD.md", "# PRD\n\n- Current-Release: 1\n\n### FR-001 a\n- Status: Active\n")
	write(t, dir, "docs/prd/b.md", "### FR-002 b\n- Status: Active\n")
	write(t, dir, "docs/other.md", "### FR-003 不是产物\n")
	s, err := Load(dir, LoadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Project.Docs) != 2 || s.Project.ByID["FR-002"] == nil || s.Project.ByID["FR-003"] != nil {
		t.Fatalf("docs: %d", len(s.Project.Docs))
	}
	if got := s.PRDDocs(); len(got) != 2 || got[0].Path != PRDIndex {
		t.Fatalf("prd docs: %v", got)
	}
}

func TestLoad_UnreadableArtifact_IF_009(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root 可以读取任何文件")
	}
	dir := t.TempDir()
	write(t, dir, "docs/PRD.md", "# PRD\n")
	if err := os.Chmod(filepath.Join(dir, "docs/PRD.md"), 0o000); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir, LoadOptions{}); err == nil {
		t.Fatal("want error")
	}
}

func TestMatch_IF_009(t *testing.T) {
	cases := []struct {
		p, f string
		ok   bool
	}{
		{"internal/**", "internal/a/b.go", true},
		{"internal/**", "internal", true},
		{"cmd/*.go", "cmd/main.go", true},
		{"cmd/*.go", "cmd/x/main.go", false},
		{"**/*_test.go", "a/b/c_test.go", true},
		{"docs/PRD.md", "docs/PRD.md", true},
		{"internal/app/**", "internal/apple/x.go", false},
	}
	for _, c := range cases {
		if Match(c.p, c.f) != c.ok {
			t.Errorf("Match(%q, %q) != %v", c.p, c.f, c.ok)
		}
	}
}

func TestLoadSourcesAndReports_IF_009(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "harness.json", `{"sourcePaths": ["internal/**"]}`)
	write(t, dir, "internal/a.go", "package a\n")
	write(t, dir, "internal/a_integration_test.go", "package a\n")
	write(t, dir, "other/b.go", "package b\n")
	write(t, dir, ".harness/reports/go.jsonl", `{"Action":"pass","Package":"m/a","Test":"TestA"}`+"\n")
	s, err := Load(dir, LoadOptions{WithSources: true, WithTests: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Sources) != 2 || s.Sources[1].Kind != 2 || len(s.Tests) != 1 || len(s.Reports) != 1 {
		t.Fatalf("sources %+v tests %+v reports %v", s.Sources, s.Tests, s.Reports)
	}
}

func TestInterfaces_IF_009(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a/a.go", "package a\n\n// harness:IF-002\nfunc B() {}\n\n// harness:IF-001\nfunc A(x int) error { return nil }\n")
	write(t, dir, "b/b.go", "package b\n\n// harness:IF-001\nfunc Other() {}\n")
	write(t, dir, "a/a_test.go", "package a\n\n// harness:IF-003\nfunc helper() {}\n")
	s, err := Load(dir, LoadOptions{WithSources: true})
	if err != nil {
		t.Fatal(err)
	}
	sigs := s.ExtractInterfaces()
	want := map[string]string{"IF-001": "func A(x int) error", "IF-002": "func B()"}
	if len(sigs) != 2 || sigs["IF-001"] != want["IF-001"] || sigs["IF-002"] != want["IF-002"] {
		t.Fatalf("同一接口取第一个锚点，测试文件不计入：%v", sigs)
	}
	out := FormatInterfaces(sigs)
	if string(out) != "IF-001\nfunc A(x int) error\n\nIF-002\nfunc B()\n" {
		t.Fatalf("格式：%q", out)
	}
	if back := ParseInterfaces(out); len(back) != 2 || back["IF-002"] != "func B()" {
		t.Fatalf("解析：%v", back)
	}
	write(t, dir, InterfacesFile, string(out))
	if got, ok := s.Interfaces(); !ok || got["IF-001"] != want["IF-001"] {
		t.Fatalf("读取快照：%v %v", got, ok)
	}
}
