package golang

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/donkeydluffy/harness/internal/lang/spec"
)

var _ spec.Analyzer = Lang{}

func TestNormalize_IF_035(t *testing.T) {
	cases := map[string]string{
		"func Check(root string, opt CheckOptions) (engine.Report, error)":                "func Check(root string, opt CheckOptions) (engine.Report, error)",
		"func  Check( root string,\n\topt CheckOptions, // 选项\n) (engine.Report,  error)": "func Check(root string, opt CheckOptions) (engine.Report, error)",
		"func (r *Rule) Check(s *workspace.Snapshot) []engine.Violation":                  "func (r *Rule) Check(s *workspace.Snapshot) []engine.Violation",
		"type Import struct{ Path string; Line int }":                                     "type Import struct { Path string; Line int }",
		"type Import struct {\n\tPath string // 路径\n\tLine int\n}":                        "type Import struct { Path string; Line int }",
		"type M map[string][]int":                                                         "type M map[string][]int",
		"func F[T any](xs ...T) func(int) <-chan T":                                       "func F[T any](xs ...T) func(int) <-chan T",
		"type L[T any] struct{}":                                                          "type L[T any] struct {}",
		"type S struct {\n\tItems  []string `json:\"items\"`\n\tM map[string]int\n}":      "type S struct { Items []string; M map[string]int }",
		"func f(xs []int, a [3]T) List[T]":                                                "func f(xs []int, a [3]T) List[T]",
		"const C = \"x\"":                                                                 "const C = \"x\"",
		"type A interface {\n\tX() /* c */ int\n\tY(a, b string)\n}":                      "type A interface { X() int; Y(a, b string) }",
	}
	for in, want := range cases {
		if got := Normalize([]byte(in)); got != want {
			t.Errorf("Normalize(%q)\n got %q\nwant %q", in, got, want)
		}
	}
}

const anchored = `package a

// Check 检查。
//
// harness:IF-022
func Check(root string) error {
	return nil // 注释
}

// harness:IF-023
type T struct {
	A int // 字段
}

// harness:IF-024

var x = 1
`

func TestAnchors_IF_035(t *testing.T) {
	got := Lang{}.Anchors("a/a.go", []byte(anchored))
	want := []spec.Decl{
		{ID: "IF-022", File: "a/a.go", Line: 5, EntryLine: 7, Signature: "func Check(root string) error"},
		{ID: "IF-023", File: "a/a.go", Line: 10, Signature: "type T struct { A int }"},
		{ID: "IF-024", File: "a/a.go", Line: 15},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
	changed := []byte("package a\n\n// 另一段说明\n// harness:IF-022\nfunc Check(root string) error { return os.ErrClosed }\n")
	if s := (Lang{}).Anchors("a.go", changed)[0].Signature; s != want[0].Signature {
		t.Fatalf("只改函数体与注释时签名不应变化：%q", s)
	}
}

func TestImports_IF_035(t *testing.T) {
	src := []byte("package a\n\nimport (\n\t\"fmt\"\n\tx \"example.com/m/b\"\n)\n")
	want := []spec.Import{{Path: "fmt", Line: 4}, {Path: "example.com/m/b", Line: 5}}
	if got := (Lang{}).Imports(src); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
}

func TestResolve_IF_035(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("// c\nmodule example.com/m\n\ngo 1.21\n"), 0o644)
	if mod, line, ok := (Lang{}).ModuleDecl("go.mod", []byte("// c\nmodule example.com/m\n")); !ok || mod != "example.com/m" || line != 2 {
		t.Fatalf("ModuleDecl：%q %d %v", mod, line, ok)
	}
	if _, _, ok := (Lang{}).ModuleDecl("a.go", nil); ok {
		t.Fatal("不是 go.mod")
	}
	for in, want := range map[string]string{"example.com/m/internal/a": "internal/a", "example.com/m": "."} {
		if got, ok := (Lang{}).Resolve(dir, in); !ok || got != want {
			t.Errorf("Resolve(%q) = %q %v", in, got, ok)
		}
	}
	for _, in := range []string{"fmt", "example.com/mx/a", "golang.org/x/tools"} {
		if _, ok := (Lang{}).Resolve(dir, in); ok {
			t.Errorf("%s 不是仓库内的包", in)
		}
	}
}

func TestCoverage_IF_035(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("没有 go")
	}
	dir := t.TempDir()
	files := map[string]string{
		"go.mod":      "module example.com/m\n\ngo 1.21\n",
		"a/a.go":      "package a\n\nfunc A(x int) int {\n\tif x > 0 {\n\t\treturn 1\n\t}\n\treturn 0\n}\n\nfunc B() {}\n",
		"a/a_test.go": "package a\n\nimport \"testing\"\n\nfunc TestA_IF_001(t *testing.T) { A(1); B() }\n\nfunc TestA_AC_001_1_N1(t *testing.T) { A(0) }\n",
		// 集成测试以 -cover 构建并运行 main：main 只在子进程中执行
		"c/main.go": "package main\n\nfunc main() {\n\tprintln(\"c\")\n}\n",
		"c/main_test.go": "package main\n\nimport (\n\t\"os\"\n\t\"os/exec\"\n\t\"path/filepath\"\n\t\"testing\"\n)\n\n" +
			"func TestRun_AC_001_1_N2(t *testing.T) {\n\tbin := filepath.Join(t.TempDir(), \"c\")\n" +
			"\tif out, err := exec.Command(\"go\", \"build\", \"-cover\", \"-o\", bin, \".\").CombinedOutput(); err != nil {\n\t\tt.Fatal(err, string(out))\n\t}\n" +
			"\tcmd := exec.Command(bin)\n\tcmd.Env = append(os.Environ(), \"GOCOVERDIR=\"+os.Getenv(\"HARNESS_COVERDIR\"))\n" +
			"\tif err := cmd.Run(); err != nil {\n\t\tt.Fatal(err)\n\t}\n}\n",
	}
	for p, c := range files {
		os.MkdirAll(filepath.Join(dir, filepath.Dir(p)), 0o755)
		os.WriteFile(filepath.Join(dir, p), []byte(c), 0o644)
	}
	passed, tests, executed, statements, err := Lang{}.Coverage(dir, "")
	if err != nil || !passed || !reflect.DeepEqual(tests, []string{"TestA_AC_001_1_N1", "TestA_IF_001", "TestRun_AC_001_1_N2"}) {
		t.Fatalf("应运行全部测试：%v %v %v", passed, tests, err)
	}
	// 空函数体 B 是语句数为 0 的块：执行过，但不算含语句的行
	if got := executed["a/a.go"]; !reflect.DeepEqual(got, []int{4, 5, 7, 10}) {
		t.Fatalf("已执行的行：%v", executed)
	}
	if got := statements["a/a.go"]; !reflect.DeepEqual(got, []int{4, 5, 7}) {
		t.Fatalf("含语句的行：%v", statements)
	}
	if got := executed["c/main.go"]; !reflect.DeepEqual(got, []int{4}) {
		t.Fatalf("子进程执行的行应合并进结果：%v", executed)
	}
	if _, tests, _, _, _ := (Lang{}).Coverage(dir, "TestA_AC"); !reflect.DeepEqual(tests, []string{"TestA_AC_001_1_N1"}) {
		t.Fatalf("pattern：%v", tests)
	}
	// 没有包声明匹配的测试时不运行 go test
	if passed, tests, executed, _, err := (Lang{}).Coverage(dir, "IF_999"); passed || tests != nil || len(executed) != 0 || err != nil {
		t.Fatalf("没有匹配的测试：%v %v %v %v", passed, tests, executed, err)
	}
	if got := testPackages(dir, "TestA_IF"); !reflect.DeepEqual(got, []string{"./a"}) {
		t.Fatalf("匹配的包：%v", got)
	}
}
