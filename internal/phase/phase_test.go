package phase

import (
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

type fakeRule struct{ id string }

func (r fakeRule) ID() string                                 { return r.id }
func (fakeRule) Covers() []string                             { return nil }
func (fakeRule) Fix() string                                  { return "x" }
func (fakeRule) Check(*workspace.Snapshot) []engine.Violation { return nil }

// 内置阶段引用的规则在本包测试中用空规则代替，真实规则由 app 注册。
func init() {
	for _, id := range []string{"prd.01", "term.01", "approve.01", "arch.01", "trace.01", "adr.01", "task.01", "task.02", "task.03", "task.04", "task.05", "task.06"} {
		engine.Register(fakeRule{id})
	}
	engine.Register(Rule())
}

func snapshot(t *testing.T, dir string) *workspace.Snapshot {
	t.Helper()
	s, err := workspace.Load(dir, workspace.LoadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestRuleSamples_IF_015(t *testing.T) {
	var samples []engine.Sample
	for _, name := range []string{"ok", "violation"} {
		samples = append(samples, engine.Sample{Name: name, Snapshot: snapshot(t, filepath.Join("testdata", "phase.01", name)), WantViolation: name == "violation"})
	}
	if err := engine.Verify(Rule(), samples); err != nil {
		t.Fatal(err)
	}
}

func TestBuiltin_IF_015(t *testing.T) {
	phases, err := Load(snapshot(t, t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, p := range phases {
		names = append(names, p.Name)
		if !p.Builtin || p.Skill == "" || len(p.Writable) == 0 || len(p.Exit) == 0 {
			t.Errorf("阶段 %s 定义不完整：%+v", p.Name, p)
		}
	}
	if strings.Join(names, ",") != "requirements,architecture,planning,implementation" || phases[3].Order != 4 {
		t.Fatalf("阶段顺序：%v", names)
	}
	if len(phases[0].Entry) != 0 {
		t.Fatal("第一个阶段的进入规则应为空")
	}
}

func TestBuiltinMissingExit_IF_015(t *testing.T) {
	fsys := fstest.MapFS{
		"builtin/order.json": {Data: []byte(`["a"]`)},
		"builtin/a.json":     {Data: []byte(`{"name":"a","entry":[],"writable":[],"skill":"x"}`)},
	}
	_, probs := load(fsys, snapshot(t, t.TempDir()))
	if len(probs) != 1 || !strings.Contains(probs[0].Msg, "缺少 exit") || probs[0].File != "internal/phase/builtin/a.json" {
		t.Fatalf("%+v", probs)
	}
}

// 演示 EXT-04：新增一个内置阶段只需新增阶段文件并在 order.json 登记一行。
func TestExample_EXT_04(t *testing.T) {
	fsys := fstest.MapFS{"builtin/review.json": {Data: []byte(`{"name":"review","entry":["task.01"],"exit":["prd.*"],"writable":["docs/review/**"],"skill":"review"}`)}}
	for _, f := range []string{"requirements", "architecture", "planning", "implementation"} {
		b, _ := builtinFS.ReadFile("builtin/" + f + ".json")
		fsys["builtin/"+f+".json"] = &fstest.MapFile{Data: b}
	}
	fsys["builtin/order.json"] = &fstest.MapFile{Data: []byte(`["requirements","architecture","planning","review","implementation"]`)}
	phases, probs := load(fsys, snapshot(t, t.TempDir()))
	if len(probs) > 0 || len(phases) != 5 || phases[3].Name != "review" || phases[3].Order != 4 || !phases[3].Builtin {
		t.Fatalf("%+v %+v", phases, probs)
	}
}

func TestSelector_IF_015(t *testing.T) {
	if sel, ok := Selector([]string{"*"}); !ok || len(sel.IDs)+len(sel.Prefixes) != 0 {
		t.Fatal("* 应为全部规则")
	}
	if _, ok := Selector(nil); ok {
		t.Fatal("空列表不运行规则")
	}
	if sel, _ := Selector([]string{"prd.*", "term.01"}); sel.Prefixes[0] != "prd" || sel.IDs[0] != "term.01" {
		t.Fatalf("%+v", sel)
	}
}
