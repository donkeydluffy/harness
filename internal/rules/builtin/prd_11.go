package builtin

import (
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// prd11：Scenario 类 NFR 写明新增什么、允许路径、封闭模块；PRD 没有 Scenario 类 NFR 时，
// 非功能需求章节须写明"不预留扩展"。
type prd11 struct{}

func (prd11) ID() string       { return "prd.11" }
func (prd11) Covers() []string { return []string{"AC-033-1", "AC-033-5"} }
func (prd11) Fix() string {
	return "为扩展场景 NFR 写 - Adds: 新增什么、- Allowed-Paths: 只允许改动的路径、- Closed-Modules: 不得修改的模块；没有扩展场景时在非功能需求章节写明\"不预留扩展\""
}

const noExtension = "不预留扩展"

func (prd11) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	scenarios := 0
	for _, e := range requirements(s) {
		if e.Kind != "NFR" || e.Fields["Type"] != "Scenario" || workspace.Removed(e) {
			continue
		}
		scenarios++
		for _, f := range []string{"Adds", "Allowed-Paths", "Closed-Modules"} {
			if strings.TrimSpace(e.Fields[f]) == "" {
				out = append(out, at(e, "扩展场景 %s 缺少 %s 字段", e.ID, f))
			}
		}
	}
	idx := s.Doc(workspace.PRDIndex)
	if scenarios > 0 || idx == nil {
		return out
	}
	for _, d := range s.PRDDocs() {
		if start, end, ok := chapter(d, "非功能需求"); ok {
			for n := start; n < end && n <= len(d.Lines); n++ {
				if strings.Contains(d.Lines[n-1], noExtension) {
					return out
				}
			}
		}
	}
	return append(out, engine.Violation{File: idx.Path, Line: 1,
		Msg: "PRD 没有 Type 为 Scenario 的非功能需求，非功能需求章节也没有写明\"不预留扩展\""})
}
