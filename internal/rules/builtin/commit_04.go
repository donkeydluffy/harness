package builtin

import (
	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// commit04：扩展任务（Extends）的提交不能改动扩展点封闭模块路径下的文件。
type commit04 struct{}

func (commit04) Needs() engine.Input { return engine.InputGit }
func (commit04) ID() string          { return "commit.04" }
func (commit04) Covers() []string    { return []string{"AC-034-1"} }
func (commit04) Fix() string {
	return "扩展任务只改扩展点的 Paths；确需修改封闭模块时先通过 ADR 调整扩展点设计"
}

func (commit04) Check(s *workspace.Snapshot) []engine.Violation {
	sets, missing := changeSets(s)
	if missing != nil {
		return []engine.Violation{*missing}
	}
	var out []engine.Violation
	for _, cs := range sets {
		t := commitTask(s, cs)
		if t == nil {
			continue
		}
		for _, x := range t.List("Extends") {
			ext := s.Project.ByID[x]
			if ext == nil {
				continue
			}
			for _, cm := range ext.List("Closed") {
				mod := s.Project.ByID[cm]
				if mod == nil {
					continue
				}
				for _, f := range cs.Files {
					if workspace.MatchAny(mod.List("Paths"), f.Path) {
						out = append(out, commitViolation(cs, f.Path, "%s 扩展 %s，改动了封闭模块 %s 的文件 %s", t.ID, x, cm, f.Path))
					}
				}
			}
		}
	}
	return out
}
