package lang

import (
	"fmt"
	"testing"

	"github.com/donkeydluffy/harness/internal/lang/spec"
)

type fakeLang struct{}

func (fakeLang) Name() string                          { return "fake" }
func (fakeLang) Extensions() []string                  { return []string{".fk"} }
func (fakeLang) Comments([]byte) []Comment             { return []Comment{{Text: "# x", Line: 1}} }
func (fakeLang) TestKind(string, []byte) spec.TestKind { return Integration }
func (fakeLang) TestNames([]byte) []string             { return []string{"t1"} }

// 演示 EXT-01：新增一种语言只需实现 Language 并注册。
func TestExample_EXT_01(t *testing.T) {
	Register(fakeLang{})
	l, ok := ForPath("a/b.FK")
	if !ok || l.Name() != "fake" || len(All()) < 2 {
		t.Fatalf("lookup: %v %v", l, ok)
	}
	if _, ok := ForPath("x.unknown"); ok {
		t.Fatal("unknown extension")
	}
}

func TestGo_IF_003(t *testing.T) {
	g, ok := ForPath("internal/a.go")
	if !ok || g.Name() != "go" {
		t.Fatal("go not registered")
	}
	src := []byte("package a\n\n// TODO(T-001): x\nvar s = \"// not a comment\"\n/* block */\n")
	cs := g.Comments(src)
	if len(cs) != 2 || cs[0].Line != 3 || cs[1].Text != "/* block */" {
		t.Fatalf("comments: %+v", cs)
	}
	if g.TestKind("a.go", nil) != NotTest || g.TestKind("a_test.go", []byte("package a")) != Unit ||
		g.TestKind("a_integration_test.go", nil) != Integration ||
		g.TestKind("b_test.go", []byte("//go:build integration\n\npackage a")) != Integration {
		t.Fatal("test kind")
	}
	names := g.TestNames([]byte("package a\nimport \"testing\"\nfunc TestA(t *testing.T){}\nfunc TestMain(m *testing.M){}\nfunc helper(){}\nfunc TestB_AC_001_1_N1(t *testing.T){}\n"))
	if len(names) != 2 || names[1] != "TestB_AC_001_1_N1" {
		t.Fatalf("names: %v", names)
	}
}

// 非 Go 语言包：注释（跳过字符串）、测试分类、测试名。
func TestLanguages_IF_003(t *testing.T) {
	type tc struct {
		file        string
		src         string
		comments    []int // 注释行号
		unit, integ string
		notTest     string
		testSrc     string
		names       []string
	}
	cases := map[string]tc{
		"typescript": {"src/a.ts", "const s = \"// no\";\n// TODO(T-1)\nconst t = `\n/* no */`;\n/* a\nb */ x\n", []int{2, 5},
			"src/a.test.ts", "tests/integration/a.test.ts", "src/a.ts",
			"test('TestA_AC_001_1_N1', () => {});\nit(\"b c\", () => {});\n", []string{"TestA_AC_001_1_N1", "b c"}},
		"python": {"app/a.py", "s = '# no'\n# FIXME T-5\nd = \"\"\"\n# no\n\"\"\"\nx = 1  # tail\n", []int{2, 6},
			"tests/test_a.py", "tests/integration/test_a.py", "app/a.py",
			"def test_ac_001_1_n1():\n    pass\nclass T:\n    async def test_b(self):\n        pass\ndef helper(): pass\n", []string{"test_ac_001_1_n1", "test_b"}},
		"java": {"src/A.kt", "val s = \"// no\"\n// TODO(T-1)\nval c = '/'\n/* x */\n", []int{2, 4},
			"src/test/ATest.java", "src/test/AIT.java", "src/A.java",
			"class ATest {\n  @Test\n  void testAc0011N1() {}\n  @Test\n  @DisplayName(\"x\")\n  fun `AC_001_1_B1 works`() {}\n}\n", []string{"testAc0011N1", "AC_001_1_B1 works"}},
		"cpp": {"src/a.cc", "auto s = \"/* no */\";\nchar c = '\\'';\n// TODO(T-1)\n", []int{3},
			"tests/a_test.cc", "tests/integration/a_test.cc", "src/a.cc",
			"TEST(Init, AC_001_1_N1) {}\nTEST_F(Fx, B) {}\n", []string{"AC_001_1_N1", "B"}},
	}
	for name, c := range cases {
		l, ok := ForPath(c.file)
		if !ok || l.Name() != name {
			t.Fatalf("%s 未注册（%s）", name, c.file)
		}
		var lines []int
		for _, cm := range l.Comments([]byte(c.src)) {
			lines = append(lines, cm.Line)
		}
		if fmt.Sprint(lines) != fmt.Sprint(c.comments) {
			t.Errorf("%s 注释行：%v，期望 %v", name, lines, c.comments)
		}
		if l.TestKind(c.unit, nil) != Unit || l.TestKind(c.integ, nil) != Integration || l.TestKind(c.notTest, nil) != NotTest {
			t.Errorf("%s 测试分类", name)
		}
		if got := l.TestNames([]byte(c.testSrc)); fmt.Sprint(got) != fmt.Sprint(c.names) {
			t.Errorf("%s 测试名：%q，期望 %q", name, got, c.names)
		}
	}
}

func TestPackFor_IF_034(t *testing.T) {
	for _, name := range []string{"go", "python", "typescript", "java", "cpp"} {
		p, ok := PackFor(name)
		if !ok || len(p.SourcePaths) == 0 || p.Format.Cmd == "" || p.Lint.Cmd == "" || p.Format.Install == "" || p.Lint.Install == "" {
			t.Errorf("%s 的语言配置包不完整：%+v", name, p)
		}
		for _, tool := range []Tool{p.Format, p.Lint} {
			if tool.Output != "files" && tool.Output != "lines" {
				t.Errorf("%s 的输出格式 %q 无效", name, tool.Output)
			}
		}
	}
	if p, _ := PackFor("go"); p.Format.Cmd != "gofmt -l ." || p.Lint.Cmd != "go vet ./..." {
		t.Errorf("Go 的默认工具：%+v", p)
	}
	if _, ok := PackFor("rust"); ok {
		t.Error("未注册的语言没有配置包")
	}
}
