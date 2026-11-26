package builtin

import (
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// commit01：每个提交恰好一个 Task 尾注，且引用存在的任务；只改产物与锁文件的提交可改用一个 Change 尾注（AC-013-5）。
// 检查暂存区且没有提交信息时不检查。
type commit01 struct{}

func (commit01) Needs() engine.Input { return engine.InputGit }
func (commit01) ID() string          { return "commit.01" }
func (commit01) Covers() []string    { return []string{"AC-013-1", "AC-013-5"} }
func (commit01) Fix() string {
	return "在提交信息末尾空一行后写一个尾注 Task: T-001，引用任务列表中存在的任务；只改 docs/ 与 harness.lock 的需求提交可改用 Change: CHG-001；不相关的改动拆成多个提交"
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
		ts, chg := cs.Commit.Trailers["Task"], cs.Commit.Trailers["Change"]
		switch {
		case len(ts) > 0 && len(chg) > 0:
			out = append(out, commitViolation(cs, "", "提交信息同时有 Task 与 Change 尾注，只能用其中一种"))
		case len(chg) > 1:
			out = append(out, commitViolation(cs, "", "提交信息有 %d 个 Change 尾注（%s），一个提交只能对应一个变更", len(chg), strings.Join(chg, ", ")))
		case len(chg) == 1:
			out = append(out, changeTrailer(s, cs, chg[0])...)
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

// changeTrailer 校验 Change 尾注：引用变更日志中存在的 CHG，且提交只改动 docs/ 下的产物与锁文件。
func changeTrailer(s *workspace.Snapshot, cs changeSet, id string) []engine.Violation {
	var out []engine.Violation
	found := false
	for _, c := range s.Project.Changes {
		found = found || c.ID == id
	}
	if !found {
		out = append(out, commitViolation(cs, "", "Change 尾注引用的 %s 不在变更日志中", id))
	}
	for _, f := range cs.Files {
		if !strings.HasPrefix(f.Path, "docs/") && f.Path != "harness.lock" {
			out = append(out, commitViolation(cs, f.Path, "带 Change 尾注的提交只能改动 docs/ 与 harness.lock，%s 需要关联任务", f.Path))
		}
	}
	return out
}
