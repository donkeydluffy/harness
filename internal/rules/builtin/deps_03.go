package builtin

import (
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// deps03：模块 Depends 中声明的模块须在本模块的非测试代码中被 import，否则是过时声明（AC-053-3）。
// 只检查路径内有可分析源文件（Release 2 为 Go）的模块。
type deps03 struct{}

func (deps03) ID() string          { return "deps.03" }
func (deps03) Covers() []string    { return []string{"AC-053-3"} }
func (deps03) Needs() engine.Input { return engine.InputSources }
func (deps03) Fix() string {
	return "从 Depends 中删去没有使用的模块（设计改动须由人重新批准），或确认依赖确实存在于非测试代码中"
}

func (deps03) Check(s *workspace.Snapshot) []engine.Violation {
	used := map[string]map[string]bool{} // 模块 -> 被 import 的模块
	for _, im := range crossImports(s) {
		if used[im.from.ID] == nil {
			used[im.from.ID] = map[string]bool{}
		}
		used[im.from.ID][im.to.ID] = true
	}
	var out []engine.Violation
	for _, m := range modules(s) {
		if !analyzable(s, m) {
			continue
		}
		for _, d := range m.List("Depends") {
			if dep := s.Project.ByID[d]; dep != nil && dep.Kind == "MOD" && !used[m.ID][d] {
				out = append(out, at(m, "%s 的 Depends 声明了 %s，但非测试代码中没有 import 它（过时声明）", m.ID, d))
			}
		}
	}
	return out
}
