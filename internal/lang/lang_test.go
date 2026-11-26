package lang

import (
	"testing"

	"harness/internal/lang/spec"
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
