package builtin

import (
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// doc01：报告其他结构规则不负责的解析诊断（如 PRD 中未结束的代码块、跨文件重复的编号、进度文件的诊断），
// 使解析问题不被静默忽略（见 diagOwner）。
type doc01 struct{}

func (doc01) ID() string       { return "doc.01" }
func (doc01) Covers() []string { return []string{"AC-003-1"} }
func (doc01) Fix() string {
	return "按 docs/architecture/formats.md 修正结构：补上代码块的结束围栏，重复的编号改为唯一编号"
}

func (doc01) Check(s *workspace.Snapshot) []engine.Violation { return ownDiags(s, "doc.01") }
