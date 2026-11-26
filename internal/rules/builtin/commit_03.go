package builtin

import (
	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// commit03：单个提交改动的行数（新增加删除，不含产物文件和生成文件）不超过 commitMaxLines。
type commit03 struct{}

func (commit03) Needs() engine.Input { return engine.InputGit }
func (commit03) ID() string          { return "commit.03" }
func (commit03) Covers() []string    { return []string{"AC-013-4"} }
func (commit03) Fix() string {
	return "把提交拆成几个各自独立的小提交（如接口、实现、测试分开）"
}

func (commit03) Check(s *workspace.Snapshot) []engine.Violation {
	sets, missing := changeSets(s)
	if missing != nil {
		return []engine.Violation{*missing}
	}
	limit := s.Config.CommitMaxLines
	var out []engine.Violation
	for _, cs := range sets {
		n := 0
		for _, f := range cs.Files {
			if !exemptFile(f.Path) {
				n += f.Added + f.Deleted
			}
		}
		if n > limit {
			out = append(out, commitViolation(cs, "", "改动 %d 行，超过上限 %d 行（不含产物文件和生成文件）", n, limit))
		}
	}
	return out
}
