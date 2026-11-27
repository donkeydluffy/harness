package builtin

import (
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// deps02：sourcePaths 内的源文件须属于某个模块的 Paths（AC-053-2）。sourcePaths 之外的文件不加载，因此不检查。
type deps02 struct{}

func (deps02) ID() string          { return "deps.02" }
func (deps02) Covers() []string    { return []string{"AC-053-2"} }
func (deps02) Needs() engine.Input { return engine.InputSources }
func (deps02) Fix() string {
	return "把文件移到某个模块的路径内，或在架构设计中扩大模块的 Paths（设计改动须由人重新批准）"
}

func (deps02) Check(s *workspace.Snapshot) []engine.Violation {
	mods := modules(s)
	if len(mods) == 0 {
		return nil // 还没有架构设计
	}
	var out []engine.Violation
	for _, f := range s.Sources {
		if owner(mods, f.Path) == nil {
			out = append(out, engine.Violation{File: f.Path, Line: 1, Msg: f.Path + " 不属于任何模块的路径（未归属文件）"})
		}
	}
	return out
}
