package custom

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/donkeydluffy/harness/internal/workspace"
)

func snapshot(t *testing.T, files map[string]string, defs ...string) *workspace.Snapshot {
	t.Helper()
	dir := t.TempDir()
	files["harness.json"] = `{"rules": [` + strings.Join(defs, ",") + `]}`
	for rel, c := range files {
		p := filepath.Join(dir, rel)
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte(c), 0o644)
	}
	s, err := workspace.Load(dir, workspace.LoadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func def(d Def) string {
	b, _ := json.Marshal(d)
	return string(b)
}

const prd = "# PRD\n\n- Status: Draft\n- Current-Release: 1\n\n### FR-001 a\n- Status: Active\n- Level: high\n- Relates: FR-009\n"

func TestLoad_IF_036(t *testing.T) {
	good := []Def{
		{ID: "acme.01", Kind: "required", Entity: "FR", Field: "Owner", Samples: "s"},
		{ID: "acme.02", Kind: "enum", Entity: "FR", Field: "Level", Values: []string{"low"}, Samples: "s"},
		{ID: "acme.03", Kind: "ref", Entity: "FR", Field: "Relates", Samples: "s"},
		{ID: "acme.04", Kind: "forbid", Paths: []string{"docs/**"}, Pattern: `FR-00\d a`, Samples: "s", ApproveFor: []string{"requirement"}},
	}
	bad := []Def{
		{ID: "acme.05", Kind: "magic", Samples: "s"},
		{ID: "prd.01", Kind: "required", Entity: "FR", Field: "Owner", Samples: "s"},
		{ID: "01", Kind: "required", Entity: "FR", Field: "Owner", Samples: "s"},
		{ID: "acme.06", Kind: "forbid", Paths: []string{"x"}, Pattern: "(", Samples: "s"},
		{ID: "acme.07", Kind: "required", Entity: "FR", Field: "Owner", Samples: "s", ApproveFor: []string{"module"}},
		{ID: "acme.08", Kind: "required", Entity: "FR", Field: "Owner"},
	}
	var defs []string
	for _, d := range append(good, bad...) {
		defs = append(defs, def(d))
	}
	s := snapshot(t, map[string]string{"docs/PRD.md": prd}, defs...)
	rules, problems := Load(s, []string{"prd.01", "arch.01"})
	if len(rules) != len(good) || len(problems) != len(bad) {
		t.Fatalf("应构造 %d 条规则、报告 %d 个问题：%d %+v", len(good), len(bad), len(rules), problems)
	}
	want := map[string]string{"acme.01": "缺少字段 Owner", "acme.02": "不在允许的取值", "acme.03": "FR-009 不存在", "acme.04": "命中禁用模式"}
	for _, r := range rules {
		vs := r.Check(s)
		if len(vs) != 1 || !strings.Contains(vs[0].Msg, want[r.ID()]) || vs[0].File != "docs/PRD.md" {
			t.Errorf("%s：%+v", r.ID(), vs)
		}
	}
	if got := ApproveChecks(rules, "requirement"); strings.Join(got, ",") != "acme.04" {
		t.Errorf("批准前检查：%v", got)
	}
	if vs := Verify(s.Root, rules[:1]); len(vs) != 1 || !strings.Contains(vs[0].Msg, "缺少样本") {
		t.Errorf("样本目录不存在时应报告：%+v", vs)
	}
}
