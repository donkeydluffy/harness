package phase

import (
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// Rule 返回本模块注册的规则 phase.01：阶段定义合法、规则引用存在、阶段名不重复、移除内置规则须引用 ADR。
func Rule() engine.Rule { return phase01{} }

type phase01 struct{}

func (phase01) ID() string       { return "phase.01" }
func (phase01) Covers() []string { return []string{"AC-045-1", "AC-045-2"} }
func (phase01) Fix() string {
	return "阶段定义须含 name、entry、exit、writable、skill；新增阶段用 after 指定位置且不与已有阶段同名；修改已有阶段用 addEntry、addExit、removeEntry、removeExit，移除时用 adr 引用 ADR"
}

func (phase01) Check(s *workspace.Snapshot) []engine.Violation {
	_, probs := load(builtinFS, s)
	var out []engine.Violation
	for _, p := range probs {
		out = append(out, engine.Violation{File: p.File, Line: 1, Msg: p.Msg})
	}
	return out
}
