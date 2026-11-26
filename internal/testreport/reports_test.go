package testreport

import (
	"path/filepath"
	"testing"
)

func TestReports_IF_033(t *testing.T) {
	dir := t.TempDir()
	if r, err := ReadLint(filepath.Join(dir, "none.json")); r != nil || err != nil {
		t.Fatalf("不存在时返回 nil：%v %v", r, err)
	}
	p := filepath.Join(dir, "r", "lint.json")
	in := &LintReport{SourceHash: "h", Issues: []LintIssue{{File: "b.go", Line: 2}, {File: "a.go", Line: 9}, {File: "a.go", Line: 1}}}
	if err := in.Write(p); err != nil {
		t.Fatal(err)
	}
	out, err := ReadLint(p)
	if err != nil || out.SourceHash != "h" || out.Issues[0] != (LintIssue{File: "a.go", Line: 1}) || out.Issues[2].File != "b.go" {
		t.Fatalf("问题应按文件、行号排序：%v %+v", err, out)
	}
	u := filepath.Join(dir, "unit.json")
	(&UnitReport{SourceHash: "h", Modules: map[string]float64{"MOD-01": 80}}).Write(u)
	if r, err := ReadUnit(u); err != nil || r.Modules["MOD-01"] != 80 || r.Skipped == nil {
		t.Fatalf("unit 报告：%v %+v", err, r)
	}
}
