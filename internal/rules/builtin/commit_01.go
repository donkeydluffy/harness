package builtin

import (
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// commit01：每个提交恰好一个 Task 尾注，且引用存在的任务。检查暂存区且没有提交信息时不检查。
type commit01 struct{}

func (commit01) Needs() engine.Input { return engine.InputGit }
func (commit01) ID() string          { return "commit.01" }
func (commit01) Covers() []string    { return []string{"AC-013-1"} }
func (commit01) Fix() string {
	return "在提交信息末尾空一行后写一个尾注 Task: T-001，引用任务列表中存在的任务；不相关的改动拆成多个提交"
}

func (commit01) Check(s *workspace.Snapshot) []engine.Violation {
	sets, missing := changeSets(s)
	if missing != nil {
		return []engine.Violation{*missing}
	}
	var out []engine.Violation
	for _, cs := range sets {
		if cs.Commit == nil {
			continue
		}
		ts := cs.Commit.Trailers["Task"]
		switch {
		case len(ts) == 0:
			out = append(out, commitViolation(cs, "", "提交信息没有 Task 尾注"))
		case len(ts) > 1:
			out = append(out, commitViolation(cs, "", "提交信息有 %d 个 Task 尾注（%s），一个提交只能对应一个任务", len(ts), strings.Join(ts, ", ")))
		case commitTask(s, cs) == nil:
			out = append(out, commitViolation(cs, "", "Task 尾注引用的 %s 不存在", ts[0]))
		}
	}
	return out
}
