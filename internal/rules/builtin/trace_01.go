package builtin

import (
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// trace01：纳入范围（Release ≤ Current-Release）的 Active FR 至少被一个有效设计元素覆盖。
type trace01 struct{}

func (trace01) ID() string       { return "trace.01" }
func (trace01) Covers() []string { return []string{"AC-007-1", "AC-003-5"} }
func (trace01) Fix() string {
	return "在覆盖该需求的模块、接口或流程的 Covers 字段中加入需求编号（Superseded 的 ADR 不计入覆盖）"
}

func (trace01) Check(s *workspace.Snapshot) []engine.Violation {
	covered := map[string]bool{}
	for _, e := range designElements(s) {
		if !effective(e) {
			continue
		}
		for _, r := range traceRefs(e) {
			covered[r] = true
		}
	}
	var out []engine.Violation
	for _, e := range requirements(s) {
		if e.Kind == "FR" && e.Fields["Status"] == "Active" && inScope(s, e) && replacedBy(s, e.ID) == "" && !covered[e.ID] {
			out = append(out, at(e, "%s 未被任何设计元素覆盖", e.ID))
		}
	}
	return out
}
