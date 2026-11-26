package builtin

import (
	"fmt"
	"sort"
	"strings"

	"harness/internal/lang"
	"harness/internal/rules/engine"
	"harness/internal/testreport"
	"harness/internal/workspace"
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
	passed, failed, other := map[string]bool{}, map[string][]string{}, map[string]string{}
	inReport := map[string]bool{}
	for _, tc := range s.Tests {
		st, ok := reportKind(tc, statics)
		if ok {
			inReport[st.File+"\x00"+st.Name] = true
		}
		for _, c := range caseRefs(tc.Name) {
			switch {
			case !ok || st.Kind != lang.Integration:
				other[c] = "只被单元测试或无法识别类型的测试引用"
			case tc.Status == testreport.Passed:
				passed[c] = true
			case tc.Status == testreport.Failed:
				failed[c] = append(failed[c], tc.Name)
			default:
				other[c] = "引用它的测试状态为 skipped"
			}
		}
	}
	var out []engine.Violation
	var ids []string
	for id := range scope {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		c := s.Project.Cases[id]
		switch {
		case len(failed[id]) > 0:
			out = append(out, engine.Violation{File: c.File, Line: c.Line,
				Msg: fmt.Sprintf("用例 %s 的集成测试失败：%s", id, strings.Join(failed[id], ", "))})
		case !passed[id]:
			why := "没有引用它的集成测试"
			if other[id] != "" {
				why = other[id]
			}
			out = append(out, engine.Violation{File: c.File, Line: c.Line,
				Msg: fmt.Sprintf("用例 %s 未被通过的集成测试覆盖（%s）", id, why)})
		}
	}
	for _, st := range statics {
		if len(caseRefs(st.Name)) > 0 && !inReport[st.File+"\x00"+st.Name] {
			out = append(out, engine.Violation{File: st.File, Line: st.Line,
				Msg: fmt.Sprintf("测试 %s 引用了用例，但没有出现在测试报告中", st.Name)})
		}
	}
	return out
}
