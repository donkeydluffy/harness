package builtin

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/donkeydluffy/harness/internal/model"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// doc04：受检 Markdown 引用的编号必须存在、不是 Removed，也没有被已生效的需求取代（AC-030-3）。
// 编号自身的定义行（标题、验收标准与用例行）以及 Replaces、Replaced-By 字段不检查。
type doc04 struct{}

func (doc04) ID() string       { return "doc.04" }
func (doc04) Covers() []string { return []string{"AC-030-3"} }
func (doc04) Fix() string {
	return "改为现行的编号（被取代的需求改为取代者），或删掉过期的引用"
}

var (
	docRefRe  = regexp.MustCompile(`(?:^|[^A-Za-z0-9-])((?:FR|NFR|IF|MOD|FLOW|ADR|T|DEBT)-\d+|AC-N?\d+-\d+(?:/[NEB]\d+)?)(?:$|[^A-Za-z0-9-])`)
	defLineRe = regexp.MustCompile(`^\s*(?:#+\s+|-\s+)((?:FR|NFR|IF|MOD|FLOW|ADR|T|DEBT)-\d+|AC-N?\d+-\d+(?:/[NEB]\d+)?)\b`)
)

func (doc04) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, l := range s.DocLines() {
		t := strings.TrimSpace(l.Text)
		if strings.HasPrefix(t, "- Replaces:") || strings.HasPrefix(t, "- Replaced-By:") {
			continue
		}
		def := ""
		if m := defLineRe.FindStringSubmatch(l.Text); m != nil {
			def = m[1]
		}
		seen := map[string]bool{}
		for _, m := range docRefRe.FindAllStringSubmatchIndex(l.Text, -1) {
			id := l.Text[m[2]:m[3]]
			if id == def || seen[id] {
				continue
			}
			seen[id] = true
			if msg := refProblem(s, id); msg != "" {
				out = append(out, engine.Violation{File: l.File, Line: l.Line, Msg: "引用的 " + id + " " + msg})
			}
		}
	}
	return out
}

// refProblem 返回编号的问题：不存在、已 Removed 或已被取代；没有问题时为空串。
func refProblem(s *workspace.Snapshot, id string) string {
	if strings.HasPrefix(id, "AC-") {
		crit, _, _ := strings.Cut(id, "/")
		it, req := caseItem(s, id)
		switch {
		case it == nil:
			return "不存在"
		case workspace.Removed(req) || (crit != id && workspace.CaseRemoved(it.Text)): // Removed 前缀只用于用例
			return "已 Removed"
		}
		for _, x := range []string{id, crit, req.ID} {
			if by := replacedBy(s, x); by != "" {
				return fmt.Sprintf("已被 %s 取代", by)
			}
		}
		return ""
	}
	kind, _, _ := strings.Cut(id, "-")
	e := s.Project.ByID[id]
	for _, x := range entityRefs(s, id) { // T、DEBT 按数值归一（T-3 与 T-003 相同）
		if x.Kind == kind {
			e = x
		}
	}
	if e != nil && e.Kind != kind {
		e = nil
	}
	switch {
	case e == nil:
		return "不存在"
	case workspace.Removed(e):
		return "已 Removed"
	}
	if by := replacedBy(s, e.ID); by != "" {
		return fmt.Sprintf("已被 %s 取代", by)
	}
	return ""
}

// caseItem 按编号查找验收标准或用例及其所属需求。
func caseItem(s *workspace.Snapshot, id string) (*model.Item, *model.Entity) {
	var find func(items []*model.Item) *model.Item
	find = func(items []*model.Item) *model.Item {
		for _, it := range items {
			if it.ID == id {
				return it
			}
			if c := find(it.Children); c != nil {
				return c
			}
		}
		return nil
	}
	for _, e := range requirements(s) {
		if it := find(e.Items); it != nil {
			return it, e
		}
	}
	return nil, nil
}
