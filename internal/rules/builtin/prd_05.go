package builtin

import (
	"strings"

	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// prd05：待澄清项状态合法，Closed 必须有 Resolution；存在 Open 的待澄清项时 PRD 不得为 Approved。
type prd05 struct{}

func (prd05) ID() string       { return "prd.05" }
func (prd05) Covers() []string { return []string{"AC-002-2"} }
func (prd05) Fix() string {
	return "先关闭待澄清项（- Status: Closed 并写 - Resolution: 结论），或把 PRD 头部的 Status 改回 Draft"
}

func (prd05) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	var open []string
	for _, d := range s.PRDDocs() {
		for _, e := range d.Entities {
			if e.Kind != "OQ" {
				continue
			}
			switch st := e.Fields["Status"]; st {
			case "Open":
				open = append(open, e.ID)
			case "Closed":
				if strings.TrimSpace(e.Fields["Resolution"]) == "" {
					out = append(out, at(e, "%s 已 Closed 但没有 Resolution 字段", e.ID))
				}
			default:
				out = append(out, at(e, "%s 的 Status 取值 %q 不合法，合法取值：Open、Closed", e.ID, st))
			}
		}
	}
	if idx := s.Doc(workspace.PRDIndex); idx != nil && idx.Header["Status"] == "Approved" && len(open) > 0 {
		out = append(out, engine.Violation{File: idx.Path, Line: 1,
			Msg: "PRD 为 Approved，但存在 Open 的待澄清项：" + strings.Join(open, ", ")})
	}
	return out
}
