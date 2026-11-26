package builtin

import (
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// prd04：PRD 包含功能需求、非功能需求、统一语言三章，且各章非空。
type prd04 struct{}

func (prd04) ID() string       { return "prd.04" }
func (prd04) Covers() []string { return []string{"AC-002-3"} }
func (prd04) Fix() string {
	return "在 docs/PRD.md 中补齐二级章节：## 统一语言（含 TERM 表格）、## 功能需求（或拆分到 docs/prd/*.md）、## 非功能需求（至少一条 NFR）"
}

func (prd04) Check(s *workspace.Snapshot) []engine.Violation {
	idx := s.Doc(workspace.PRDIndex)
	if idx == nil {
		return nil // 缺少索引由 prd.02 报告
	}
	var out []engine.Violation
	miss := func(name string) {
		out = append(out, engine.Violation{File: idx.Path, Line: 1, Msg: "PRD 缺少章节：" + name})
	}
	empty := func(name string, line int) {
		out = append(out, engine.Violation{File: idx.Path, Line: line, Msg: "章节 " + name + " 没有任何条目"})
	}
	if st, end, ok := chapter(idx, "统一语言"); !ok {
		miss("统一语言")
	} else if !anyRow(idx, st, end, "TERM") {
		empty("统一语言", st)
	}
	if st, end, ok := chapter(idx, "非功能需求"); !ok {
		miss("非功能需求")
	} else if !anyEntity(idx, st, end, "NFR") {
		empty("非功能需求", st)
	}
	split := len(s.PRDDocs()) > 1
	if st, end, ok := chapter(idx, "功能需求"); !ok && !split {
		miss("功能需求")
	} else if len(s.Project.ByKind["FR"]) == 0 || (ok && !split && !anyEntity(idx, st, end, "FR")) {
		empty("功能需求", st)
	}
	return out
}
