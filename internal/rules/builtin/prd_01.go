package builtin

import (
	"fmt"
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// prd01：PRD 编号格式合法、在全部 PRD 文件中唯一；拆分文件只能包含功能需求。
type prd01 struct{}

func (prd01) ID() string       { return "prd.01" }
func (prd01) Covers() []string { return []string{"AC-003-1", "AC-003-4"} }
func (prd01) Fix() string {
	return "编号格式为 <前缀>-<至少三位数字>（如 FR-001），且在 docs/PRD.md 与 docs/prd/*.md 中只定义一次"
}

type loc struct {
	file string
	line int
}

func (prd01) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	seen := map[string][]loc{}
	var order []string
	add := func(id, file string, line int) {
		if _, ok := seen[id]; !ok {
			order = append(order, id)
		}
		seen[id] = append(seen[id], loc{file, line})
	}
	badFormat := func(kind, id, file string, line int) {
		out = append(out, engine.Violation{File: file, Line: line,
			Msg: fmt.Sprintf("编号 %s 格式不合法，应为 %s-001 形式（至少三位数字）", id, kind)})
	}
	for _, d := range s.PRDDocs() {
		index := d.Path == workspace.PRDIndex
		for _, e := range d.Entities {
			re, ok := idFormat[e.Kind]
			if !ok {
				continue
			}
			if !re.MatchString(e.ID) {
				badFormat(e.Kind, e.ID, e.File, e.Line)
			}
			if !index && e.Kind != "FR" {
				out = append(out, at(e, "%s 只能定义在 %s 中，拆分文件只放功能需求", e.ID, workspace.PRDIndex))
			}
			add(e.ID, e.File, e.Line)
		}
		for _, t := range d.Tables {
			for _, r := range t.Rows {
				if len(r.Cells) == 0 {
					continue
				}
				id := r.Cells[0]
				for _, kind := range []string{"TERM", "CHG"} {
					if !strings.HasPrefix(id, kind) {
						continue
					}
					if !idFormat[kind].MatchString(id) {
						badFormat(kind, id, d.Path, r.Line)
					}
					if !index {
						name := map[string]string{"TERM": "统一语言", "CHG": "变更日志"}[kind]
						out = append(out, engine.Violation{File: d.Path, Line: r.Line,
							Msg: fmt.Sprintf("%s只能写在 %s 中（%s）", name, workspace.PRDIndex, id)})
					}
					// 同一 CHG 编号可占多行，只检查术语编号唯一。
					if kind == "TERM" {
						add(id, d.Path, r.Line)
					}
				}
			}
		}
	}
	for _, id := range order {
		locs := seen[id]
		if len(locs) < 2 {
			continue
		}
		var where []string
		for _, l := range locs {
			where = append(where, fmt.Sprintf("%s:%d", l.file, l.line))
		}
		for _, l := range locs[1:] {
			out = append(out, engine.Violation{File: l.file, Line: l.line,
				Msg: fmt.Sprintf("编号 %s 重复定义：%s", id, strings.Join(where, "、"))})
		}
	}
	return out
}
