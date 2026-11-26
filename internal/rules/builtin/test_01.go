package builtin

import (
	"fmt"

	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// test01：测试名中引用的用例编号必须存在。
type test01 struct{}

func (test01) Needs() engine.Input { return engine.InputSources | engine.InputTests }
func (test01) ID() string          { return "test.01" }
func (test01) Covers() []string    { return []string{"AC-020-1"} }
func (test01) Fix() string {
	return "修正测试名中的用例编号（AC-001-1/N1 写作 AC_001_1_N1），或在 PRD 中补充该用例"
}

func (test01) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	reported := map[string]bool{}
	for _, st := range staticTests(s) {
		for _, c := range caseRefs(st.Name) {
			if _, ok := s.Project.Cases[c]; !ok {
				out = append(out, engine.Violation{File: st.File, Line: st.Line,
					Msg: fmt.Sprintf("测试 %s 引用的用例 %s 不存在", st.Name, c)})
			}
		}
		reported[st.Name] = true
	}
	for _, tc := range s.Tests {
		for _, c := range caseRefs(tc.Name) {
			if _, ok := s.Project.Cases[c]; !ok && !reported[tc.Name] {
				out = append(out, engine.Violation{File: s.Reports[0], Line: 1,
					Msg: fmt.Sprintf("测试报告中的 %s 引用的用例 %s 不存在", tc.Name, c)})
			}
		}
	}
	return out
}
