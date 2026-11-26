package builtin

import (
	"fmt"
	"regexp"

	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// debt02：源码中的待办标记注释必须在同一行引用 Open 的技术债或未完成的任务。
type debt02 struct{}

func (debt02) Needs() engine.Input { return engine.InputSources }
func (debt02) ID() string          { return "debt.02" }
func (debt02) Covers() []string    { return []string{"AC-031-2"} }
func (debt02) Fix() string {
	return "在同一行引用 Open 的技术债或未完成的任务，如 // TODO(DEBT-002): 重试；没有对应项时先在 docs/debt.md 登记"
}

var markerRe = regexp.MustCompile(`(?:^|[^A-Za-z0-9_])(TODO|FIXME|HACK|XXX)(?:[^A-Za-z0-9_]|$)`)

func (debt02) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, c := range commentLines(s) {
		m := markerRe.FindStringSubmatch(c.Text)
		if m == nil {
			continue
		}
		refs := entityRefs(s, c.Text)
		ok := false
		var closed []string
		for _, e := range refs {
			st := e.Fields["Status"]
			if (e.Kind == "DEBT" && st == "Open") || (e.Kind == "T" && st != "Done") {
				ok = true
			} else {
				closed = append(closed, fmt.Sprintf("%s（%s）", e.ID, st))
			}
		}
		switch {
		case ok:
		case len(closed) > 0:
			out = append(out, engine.Violation{File: c.File, Line: c.Line,
				Msg: fmt.Sprintf("%s 注释引用的 %v 已关闭，应引用 Open 的技术债或未完成的任务", m[1], closed)})
		default:
			out = append(out, engine.Violation{File: c.File, Line: c.Line,
				Msg: fmt.Sprintf("%s 注释没有引用存在的技术债或任务", m[1])})
		}
	}
	return out
}
