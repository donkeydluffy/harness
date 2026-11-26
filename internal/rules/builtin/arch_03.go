package builtin

import (
	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// arch03：调用流程每一步"调用方模块 -> 接口"中，接口存在，且接口所属模块是调用方自身或在调用方的 Depends 中。
type arch03 struct{}

func (arch03) ID() string       { return "arch.03" }
func (arch03) Covers() []string { return []string{"AC-006-5"} }
func (arch03) Fix() string {
	return "修正流程步骤，或在调用方模块的 Depends 中加入接口所属模块（依赖变化需 ADR 引用，见 adr.03）"
}

func (arch03) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, f := range archEntities(s, "FLOW") {
		steps, _ := parseSteps(f.Fields["Steps"])
		for i, st := range steps {
			if st.Caller == "" {
				continue // 格式错误由 arch.01 报告
			}
			caller := s.Project.ByID[st.Caller]
			if caller == nil || caller.Kind != "MOD" {
				out = append(out, at(f, "%s 第 %d 步的调用方 %s 不存在", f.ID, i+1, st.Caller))
				continue
			}
			target := s.Project.ByID[st.IF]
			if target == nil || target.Kind != "IF" {
				out = append(out, at(f, "%s 第 %d 步调用的接口 %s 不存在", f.ID, i+1, st.IF))
				continue
			}
			owner := target.Fields["Module"]
			if owner != st.Caller && !contains(caller.List("Depends"), owner) {
				out = append(out, at(f, "%s 第 %d 步：%s 调用 %s，但其所属模块 %s 不在 %s 的 Depends 中",
					f.ID, i+1, st.Caller, st.IF, owner, st.Caller))
			}
		}
	}
	return out
}
