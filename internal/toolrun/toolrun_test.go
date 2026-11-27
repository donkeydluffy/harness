package toolrun

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
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

const (
	unitOverview = "# 架构总览\n\n### MOD-01 a\n- Status: Active\n- Paths: internal/a/**\n- Public: internal/a/a.go\n- Depends:\n- Interfaces: IF-001, IF-002\n- Covers: FR-001\n- Core: false\n- Coverage: 60\n"
	unitModules  = "# 模块详细设计\n\n### IF-001 A\n- Module: MOD-01\n- Covers: FR-001\n\n### IF-002 B\n- Module: MOD-01\n- Covers: FR-001\n"
	unitGo       = "package a\n\n// harness:IF-001\nfunc A(x int) int {\n\tif x > 0 {\n\t\treturn 1\n\t}\n\treturn 0\n}\n\n// harness:IF-002\nfunc B() int {\n\treturn 2\n}\n"
	// TestB_IF_002 只调用了 A：IF-002 的入口行（第 13 行）未执行
	unitTest = "package a\n\nimport \"testing\"\n\nfunc TestA_IF_001(t *testing.T) { A(1) }\n\nfunc TestB_IF_002(t *testing.T) { A(0) }\n"
)

func TestUnit_IF_031(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("没有 go")
	}
	s := snapshot(t, map[string]string{
		"harness.json": `{"languages": ["go", "python"], "sourcePaths": ["internal/**"]}`, "go.mod": "module demo\n\ngo 1.21\n",
		"docs/architecture/overview.md": unitOverview, "docs/architecture/modules/a.md": unitModules,
		"internal/a/a.go": unitGo, "internal/a/a_test.go": unitTest, "internal/b/b.py": "x = 1\n",
	})
	rep, err := Unit(s)
	if err != nil || rep.SourceHash != s.SourceHash() || strings.Join(rep.Skipped, ",") != "python" {
		t.Fatalf("没有 Analyzer 的语言记为 Skipped：%v %+v", err, rep)
	}
	want := map[string]testreport.IFCoverage{
		"IF-001": {Tests: []string{"TestA_IF_001"}, Passed: true, EntryHit: true},
		"IF-002": {Tests: []string{"TestB_IF_002"}, Passed: true, EntryHit: false},
	}
	if !reflect.DeepEqual(rep.Interfaces, want) {
		t.Fatalf("接口：%+v", rep.Interfaces)
	}
	// 含语句的行 5、6、8、13，全量运行执行了前 3 行
	if got := rep.Modules["MOD-01"]; got != 75 {
		t.Fatalf("模块覆盖率：%v", rep.Modules)
	}
	if r, _ := testreport.ReadUnit(filepath.Join(s.Root, testreport.UnitFile)); r == nil || !reflect.DeepEqual(r.Interfaces, want) {
		t.Fatal("应写入 unit 报告")
	}
}

func TestTestPattern(t *testing.T) {
	re := regexp.MustCompile(testPattern("IF-002"))
	for name, want := range map[string]bool{"TestSave_IF_002": true, "TestSave_IF_002_Empty": true, "TestSave_IF_0021": false, "TestSave_IF_001": false} {
		if re.MatchString(name) != want {
			t.Errorf("%s 应为 %v", name, want)
		}
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
