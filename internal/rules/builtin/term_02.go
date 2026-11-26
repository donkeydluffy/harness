package builtin

import (
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// term02：架构声明中以 Domain 字段标注的领域概念必须出现在统一语言表格的"代码名"列。
type term02 struct{}

func (term02) ID() string       { return "term.02" }
func (term02) Covers() []string { return []string{"AC-005-2"} }
func (term02) Fix() string {
	return "在 PRD 统一语言表格中登记该领域概念（代码名列），或改用已登记的代码名"
}

func (term02) Check(s *workspace.Snapshot) []engine.Violation {
	codes := map[string]bool{}
	for _, t := range s.Project.Terms {
		codes[t.Code] = true
	}
	var out []engine.Violation
	for _, d := range s.Project.Docs {
		if !strings.HasPrefix(d.Path, "docs/architecture/") {
			continue
		}
		for _, e := range d.Entities {
			for _, name := range e.List("Domain") {
				if !codes[name] {
					out = append(out, at(e, "%s 标注的领域概念 %s 不在统一语言表格的代码名列中", e.ID, name))
				}
			}
		}
	}
	return out
}
