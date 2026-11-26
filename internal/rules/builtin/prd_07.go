package builtin

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// prd07：用例编号格式合法、前缀与所在验收标准一致；纳入范围的 Active 需求的每条验收标准
// 至少一个 N 用例和一个 B 或 E 用例。同时报告解析阶段的用例诊断。
type prd07 struct{}

func (prd07) ID() string       { return "prd.07" }
func (prd07) Covers() []string { return []string{"AC-019-1", "AC-019-2"} }
func (prd07) Fix() string {
	return "在验收标准下缩进两格列出用例 \"  - AC-001-1/N1: …\"，类别只能是 N、B、E，每条验收标准至少一个 N 和一个 B 或 E"
}

var caseIDRe = regexp.MustCompile(`^(AC-N?\d+-\d+)/([A-Z])(\d+)$`)

func (prd07) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, e := range requirements(s) {
		check := e.Fields["Status"] == "Active" && inScope(s, e)
		for _, it := range e.Items {
			var n, be bool
			for _, c := range it.Children {
				m := caseIDRe.FindStringSubmatch(c.ID)
				v := func(msg string) {
					out = append(out, engine.Violation{File: e.File, Line: c.Line, Msg: msg})
				}
				switch {
				case m == nil:
					v(fmt.Sprintf("用例编号 %s 格式不合法，应为 <验收标准编号>/<N|B|E><序号>", c.ID))
					continue
				case !strings.Contains("NBE", m[2]):
					v(fmt.Sprintf("用例 %s 的类别 %s 不合法，只能是 N（正常）、B（边界）、E（错误）", c.ID, m[2]))
					continue
				case m[1] != it.ID:
					v(fmt.Sprintf("用例 %s 的前缀与所在验收标准 %s 不一致", c.ID, it.ID))
				}
				if workspace.CaseRemoved(c.Text) {
					continue
				}
				n = n || m[2] == "N"
				be = be || m[2] == "B" || m[2] == "E"
			}
			if check && (!n || !be) {
				out = append(out, engine.Violation{File: e.File, Line: it.Line,
					Msg: fmt.Sprintf("验收标准 %s 需要至少一个 N 用例和一个 B 或 E 用例", it.ID)})
			}
		}
	}
	for _, d := range s.Diags {
		if strings.Contains(d.Msg, "用例") {
			out = append(out, engine.Violation{File: d.File, Line: d.Line, Msg: d.Msg})
		}
	}
	return out
}
