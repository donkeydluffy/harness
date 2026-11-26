package builtin

import (
	"regexp"
	"strings"

	"harness/internal/model"
	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// debt01：技术债清单中每项有合法编号、所属模块（存在的 MOD）、描述、状态（Open / Resolved）。
type debt01 struct{}

func (debt01) ID() string       { return "debt.01" }
func (debt01) Covers() []string { return []string{"AC-031-1"} }
func (debt01) Fix() string {
	return "技术债项写作 \"### DEBT-001 标题\"，字段 - Status: Open|Resolved 与 - Module: <存在的 MOD-nn>，并在字段后写一段描述"
}

const debtFile = "docs/debt.md"

var debtIDRe = regexp.MustCompile(`^DEBT-\d{3,}$`)

func (debt01) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, e := range s.Project.ByKind["DEBT"] {
		if !debtIDRe.MatchString(e.ID) {
			out = append(out, at(e, "编号 %s 格式不合法，应为 DEBT-001 形式", e.ID))
		}
		if st := e.Fields["Status"]; !oneOf(st, "Open", "Resolved") {
			out = append(out, at(e, "%s 的 Status 取值 %q 不合法，合法取值：Open、Resolved", e.ID, st))
		}
		switch m := strings.TrimSpace(e.Fields["Module"]); {
		case m == "":
			out = append(out, at(e, "%s 缺少 Module（所属模块）", e.ID))
		case !isKind(s.Project, m, "MOD"):
			out = append(out, at(e, "%s 的所属模块 %s 不存在", e.ID, m))
		}
		if lines, _ := bodyLines(e); !anyNonBlank(lines) {
			out = append(out, at(e, "%s 缺少描述", e.ID))
		}
	}
	for _, dg := range s.Diags {
		if dg.File == debtFile {
			out = append(out, engine.Violation{File: dg.File, Line: dg.Line, Msg: dg.Msg})
		}
	}
	return out
}

func isKind(p *model.Project, id, kind string) bool {
	e := p.ByID[id]
	return e != nil && e.Kind == kind
}

func anyNonBlank(lines []string) bool {
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			return true
		}
	}
	return false
}
