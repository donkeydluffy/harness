package toolrun

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/donkeydluffy/harness/internal/lang"
	"github.com/donkeydluffy/harness/internal/testreport"
	"github.com/donkeydluffy/harness/internal/workspace"
)

func snapshot(t *testing.T, files map[string]string) *workspace.Snapshot {
	t.Helper()
	dir := t.TempDir()
	for rel, c := range files {
		p := filepath.Join(dir, rel)
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte(c), 0o644)
	}
	s, err := workspace.Load(dir, workspace.LoadOptions{WithSources: true})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestLint_IF_030(t *testing.T) {
	files := parse(lang.Tool{Cmd: "gofmt -l .", Output: "files"}, "./a.go\nWould reformat: b.py\n2 files would be reformatted\n", 0)
	if len(files) != 2 || files[0].File != "a.go" || files[1].File != "b.py" || files[0].Tool != "gofmt" {
		t.Errorf("files 格式：%+v", files)
	}
	lines := parse(lang.Tool{Cmd: "go vet ./...", Output: "lines"}, "# demo/a\nvet: ./a/c.go:5:2: bad arg\n", 1)
	if len(lines) != 1 || lines[0] != (testreport.LintIssue{File: "a/c.go", Line: 5, Tool: "go vet", Msg: "bad arg"}) {
		t.Errorf("lines 格式：%+v", lines)
	}
	failed := parse(lang.Tool{Cmd: "eslint .", Output: "lines"}, "Oops! Something went wrong\n", 2)
	if len(failed) != 1 || failed[0].File != "" || !strings.Contains(failed[0].Msg, "退出码 2") || !strings.Contains(failed[0].Msg, "Oops") {
		t.Errorf("退出码非 0 而没有解析出问题时应记一条问题：%+v", failed)
	}
	if ok := parse(lang.Tool{Cmd: "eslint .", Output: "lines"}, "", 0); len(ok) != 0 {
		t.Errorf("没有输出时没有问题：%+v", ok)
	}
	s := snapshot(t, map[string]string{"harness.json": `{"languages": ["python"], "tools": {"python": {"format": "true", "lint": "no-such-tool-x ."}}}`, "a.py": "x = 1\n"})
	if _, err := Lint(s); err == nil || err.(*MissingTool).Command != "no-such-tool-x" {
		t.Errorf("工具缺失时返回 MissingTool：%v", err)
	}
}

func TestUnit_IF_031(t *testing.T) {
	s := snapshot(t, map[string]string{"harness.json": `{"languages": ["go"]}`, "a.go": "package a\n"})
	rep, err := Unit(s)
	if err != nil || strings.Join(rep.Skipped, ",") != "go" || rep.SourceHash != s.SourceHash() {
		t.Fatalf("Release 2 尚未接入覆盖率，全部语言记为 Skipped：%v %+v", err, rep)
	}
	if r, _ := testreport.ReadUnit(filepath.Join(s.Root, testreport.UnitFile)); r == nil {
		t.Fatal("应写入 unit 报告")
	}
}

func TestSmoke_IF_032(t *testing.T) {
	if _, _, err := Smoke(snapshot(t, map[string]string{})); err == nil {
		t.Error("没有声明冒烟命令时返回错误")
	}
	code, tail, err := Smoke(snapshot(t, map[string]string{"harness.json": `{"smoke": "for i in $(seq 1 60); do echo line$i; done; exit 3"}`}))
	if err != nil || code != 3 || len(tail) != 50 || tail[49] != "line60" {
		t.Fatalf("应返回退出码与最后 50 行：%v %d %v", err, code, tail)
	}
}
