package builtin

import (
	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// commit02：提交改动的文件都在 Task 尾注所指任务的文件范围内；产物文件和生成文件除外。
type commit02 struct{}

func (commit02) Needs() engine.Input { return engine.InputGit }
func (commit02) ID() string          { return "commit.02" }
func (commit02) Covers() []string    { return []string{"AC-013-2"} }
func (commit02) Fix() string {
	return "把范围外的改动拆到对应任务的提交中，或先在 docs/tasks.md 中扩大该任务的 Scope"
}

func (commit02) Check(s *workspace.Snapshot) []engine.Violation {
	sets, missing := changeSets(s)
	if missing != nil {
		return []engine.Violation{*missing}
	}
	var out []engine.Violation
	for _, cs := range sets {
		t := commitTask(s, cs)
		if t == nil {
			continue // commit.01 报告
		}
		for _, f := range cs.Files {
			if !exemptFile(f.Path) && !workspace.MatchAny(t.List("Scope"), f.Path) {
				out = append(out, commitViolation(cs, f.Path, "%s 不在 %s 的文件范围内", f.Path, t.ID))
			}
		}
	}
	return out
}
