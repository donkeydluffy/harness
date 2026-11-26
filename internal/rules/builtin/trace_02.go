package builtin

import (
	"regexp"

	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// trace02：每个设计元素至少引用一条需求或一条 ADR，否则为孤儿。
type trace02 struct{}

func (trace02) ID() string       { return "trace.02" }
func (trace02) Covers() []string { return []string{"AC-007-2"} }
func (trace02) Fix() string {
	return "在设计元素的 Covers（模块、接口、流程）或 Relates（扩展点、ADR）中引用它服务的需求或 ADR；确实无用的元素应删除"
}

var reqOrADRRe = regexp.MustCompile(`^(FR|NFR)-\d+$|^AC-N?\d+-\d+$|^ADR-\d+$`)

func (trace02) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, e := range designElements(s) {
		ok := false
		for _, r := range traceRefs(e) {
			if reqOrADRRe.MatchString(r) {
				ok = true
				break
			}
		}
		if !ok {
			out = append(out, at(e, "设计元素 %s 没有引用任何需求或 ADR（孤儿）", e.ID))
		}
	}
	return out
}
