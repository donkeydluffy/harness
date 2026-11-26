package builtin

import (
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// prd02：需求必填字段与取值（Status、Release、Priority，NFR 另有 Type），
// Active 需求至少一条验收标准，PRD 索引头部声明 Current-Release。
type prd02 struct{}

func (prd02) ID() string { return "prd.02" }
func (prd02) Covers() []string {
	return []string{"AC-003-2", "AC-003-5", "AC-040-2"}
}
func (prd02) Fix() string {
	return "在需求标题下补齐字段：- Status: Active|Removed、- Release: <正整数>、- Priority: P0|P1|P2（NFR 另加 - Type: Quantitative|Constraint|Scenario）"
}

func (prd02) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	idx := s.Doc(workspace.PRDIndex)
	if idx == nil {
		return []engine.Violation{{File: workspace.PRDIndex, Line: 1, Msg: "缺少 PRD 索引文件",
			Fix: "创建 docs/PRD.md，或运行 harness init 生成骨架"}}
	}
	if !positiveInt(idx.Header["Current-Release"]) {
		out = append(out, engine.Violation{File: idx.Path, Line: 1,
			Msg: "PRD 头部缺少 Current-Release 或取值不是正整数",
			Fix: "在 PRD 标题下的文件头字段中加入 - Current-Release: 1"})
	}
	for _, e := range requirements(s) {
		status, ok := e.Field("Status")
		switch {
		case !ok:
			out = append(out, at(e, "%s 缺少 Status 字段", e.ID))
		case !oneOf(status, "Active", "Removed"):
			out = append(out, at(e, "%s 的 Status 取值 %q 不合法，合法取值：Active、Removed", e.ID, status))
		}
		if p, ok := e.Field("Priority"); !ok {
			out = append(out, at(e, "%s 缺少 Priority 字段", e.ID))
		} else if !oneOf(p, "P0", "P1", "P2") {
			out = append(out, at(e, "%s 的 Priority 取值 %q 不合法，合法取值：P0、P1、P2", e.ID, p))
		}
		if r, ok := e.Field("Release"); !ok {
			out = append(out, at(e, "%s 缺少 Release 字段", e.ID))
		} else if !positiveInt(r) {
			out = append(out, at(e, "%s 的 Release 取值 %q 不是正整数", e.ID, r))
		}
		if e.Kind == "NFR" {
			if t, ok := e.Field("Type"); !ok {
				out = append(out, at(e, "%s 缺少 Type 字段", e.ID))
			} else if !oneOf(t, "Quantitative", "Constraint", "Scenario") {
				out = append(out, at(e, "%s 的 Type 取值 %q 不合法，合法取值：Quantitative、Constraint、Scenario", e.ID, t))
			}
		}
		if status == "Active" && !hasAC(e.Items) {
			out = append(out, at(e, "Active 需求 %s 没有验收标准", e.ID))
		}
	}
	return out
}
