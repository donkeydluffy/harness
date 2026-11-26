package builtin

import (
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// arch04：扩展点必须关联 Scenario 类 NFR，否则是过度设计。
type arch04 struct{}

func (arch04) ID() string       { return "arch.04" }
func (arch04) Covers() []string { return []string{"AC-033-3"} }
func (arch04) Fix() string {
	return "在 PRD 中把这项可扩展性写成 Type: Scenario 的 NFR 并在扩展点的 Relates 中引用；没有对应的变化场景时删除该扩展点"
}

func (arch04) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, e := range s.Project.ByKind["EXT"] {
		if workspace.Removed(e) {
			continue
		}
		var bad []string
		ok := false
		for _, r := range e.List("Relates") {
			t := s.Project.ByID[r]
			switch {
			case t == nil: // 不存在的引用由 trace.03 报告
				ok = true
			case t.Kind == "NFR" && t.Fields["Type"] == "Scenario":
				ok = true
			default:
				bad = append(bad, r)
			}
		}
		switch {
		case len(bad) > 0:
			out = append(out, at(e, "扩展点 %s 关联的 %s 不是 Scenario 类 NFR", e.ID, strings.Join(bad, ", ")))
		case !ok:
			out = append(out, at(e, "扩展点 %s 没有关联任何 Scenario 类 NFR，属于过度设计", e.ID))
		}
	}
	return out
}
