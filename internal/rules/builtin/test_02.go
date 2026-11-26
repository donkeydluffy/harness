package builtin

import (
	"fmt"
	"sort"
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// test02：纳入范围的 Active 用例至少有一个 passed 的集成测试，且引用它的集成测试没有失败；
// skipped 与单元测试不计入覆盖；源码中声明了用例引用的测试必须出现在报告中。
type test02 struct{}

func (test02) Needs() engine.Input { return engine.InputSources | engine.InputTests }
func (test02) ID() string          { return "test.02" }
func (test02) Covers() []string {
	return []string{"AC-020-2", "AC-020-3", "AC-020-4"}
}
func (test02) Fix() string {
	return "为用例编写集成测试（测试名含用例编号，如 TestInit_AC_001_1_N1），运行后把报告写到 harness.json 的 testReports 路径，如 go test -json ./... > .harness/reports/go.jsonl"
}

func (test02) Check(s *workspace.Snapshot) []engine.Violation {
	scope := map[string]bool{}
	for _, e := range requirements(s) {
		if e.Fields["Status"] != "Active" || !inScope(s, e) {
			continue
		}
		for _, it := range e.Items {
			for _, c := range it.Children {
				if !workspace.CaseRemoved(c.Text) {
					scope[c.ID] = true
				}
			}
		}
	}
	if len(scope) == 0 {
		return nil
	}
	if len(s.Reports) == 0 {
		return []engine.Violation{{File: ".", Line: 1, Msg: "输入缺失：没有测试报告（harness.json 的 testReports：" + strings.Join(s.Config.TestReports, ", ") + "）"}}
	}
	statics := staticTests(s)
	r := caseResults(s, statics)
	var out []engine.Violation
	var ids []string
	for id := range scope {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if v := r.violation(s, id); v != nil {
			out = append(out, *v)
		}
	}
	for _, st := range statics {
		if len(caseRefs(st.Name)) > 0 && !r.inReport[st.File+"\x00"+st.Name] {
			out = append(out, engine.Violation{File: st.File, Line: st.Line,
				Msg: fmt.Sprintf("测试 %s 引用了用例，但没有出现在测试报告中", st.Name)})
		}
	}
	return out
}
