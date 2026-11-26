package builtin

import (
	"regexp"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// trace03：设计元素中的引用必须指向存在且有效的编号；引用 Superseded 的 ADR 时提示改引取代它的 ADR。
type trace03 struct{}

func (trace03) ID() string       { return "trace.03" }
func (trace03) Covers() []string { return []string{"AC-007-3"} }
func (trace03) Fix() string {
	return "修正引用：指向存在的编号；需求已 Removed 时去掉引用；ADR 已 Superseded 时改为引用取代它的 ADR"
}

// 引用字段，Superseded-By 单独检查存在性（adr.02）。
var refFields = []string{"Covers", "Relates", "Depends", "Interfaces", "Closed", "Interface", "Module"}

var refIDRe = regexp.MustCompile(`^(FR|NFR|MOD|IF|FLOW|EXT|ADR)-\d+$|^AC-N?\d+-\d+$`)

func (trace03) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, e := range designElements(s) {
		for _, f := range refFields {
			for _, r := range e.List(f) {
				if !refIDRe.MatchString(r) {
					continue // 非编号文字由各自的字段规则检查
				}
				t := target(s, r)
				if by := replacedBy(s, r); by != "" && t != "" {
					out = append(out, at(e, "%s 的 %s 引用了已被 %s 取代的 %s", e.ID, f, by, r))
					continue
				}
				switch {
				case t == "":
					out = append(out, at(e, "%s 的 %s 引用了不存在的 %s", e.ID, f, r))
				case t == "Removed":
					out = append(out, at(e, "%s 的 %s 引用了已 Removed 的 %s", e.ID, f, r))
				case t == "Superseded":
					by := s.Project.ByID[r].Fields["Superseded-By"]
					out = append(out, at(e, "%s 的 %s 引用了已 Superseded 的 %s，应改为引用 %s", e.ID, f, r, by))
				}
			}
		}
	}
	return out
}

// target 返回被引用编号的状态：不存在为空串，否则为 Status（验收标准视为 Active）。
func target(s *workspace.Snapshot, id string) string {
	if e, ok := s.Project.ByID[id]; ok {
		if st := e.Fields["Status"]; st != "" {
			return st
		}
		return "Active"
	}
	for _, e := range requirements(s) {
		for _, it := range e.Items {
			if it.ID == id {
				if workspace.Removed(e) {
					return "Removed"
				}
				return "Active"
			}
		}
	}
	return ""
}
