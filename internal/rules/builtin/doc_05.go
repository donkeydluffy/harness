package builtin

import (
	"fmt"
	"path/filepath"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// doc05：模块路径下的代码在其详细设计最后一次修改之后被提交改动超过 staleCommits 次时告警（AC-030-4）。
type doc05 struct{}

func (doc05) ID() string          { return "doc.05" }
func (doc05) Covers() []string    { return []string{"AC-030-4"} }
func (doc05) Needs() engine.Input { return engine.InputGit }
func (doc05) Fix() string {
	return "对照代码更新模块的详细设计；内容仍然正确时也可提交一次说明性的修改"
}

func (doc05) Check(s *workspace.Snapshot) []engine.Violation {
	if s.Repo == nil {
		return nil
	}
	var out []engine.Violation
	for _, m := range modules(s) {
		docs, _ := filepath.Glob(filepath.Join(s.Root, "docs", "architecture", "modules", m.ID+"-*.md"))
		if len(docs) != 1 {
			continue
		}
		rel, _ := filepath.Rel(s.Root, docs[0])
		rel = filepath.ToSlash(rel)
		last, err := s.Repo.LastCommit([]string{rel})
		if err != nil || last == "" {
			continue
		}
		var specs []string
		for _, g := range m.List("Paths") {
			specs = append(specs, ":(glob)"+g)
		}
		if len(specs) == 0 {
			continue
		}
		n, err := s.Repo.CommitsTouching(specs, last)
		if err != nil || n <= s.Config.StaleCommits {
			continue
		}
		out = append(out, engine.Violation{File: rel, Line: 1, Warn: true,
			Msg: fmt.Sprintf("%s 的代码在详细设计最后一次修改后被 %d 个提交改动（上限 %d），设计可能过期", m.ID, n, s.Config.StaleCommits)})
	}
	return out
}
