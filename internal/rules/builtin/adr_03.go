package builtin

import (
	"fmt"
	"sort"
	"strings"

	"harness/internal/model"
	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// adr03：同一提交中，架构声明的模块增删、允许依赖变化、扩展点变化必须由该提交新增或修改的 ADR 引用。
type adr03 struct{}

func (adr03) ID() string       { return "adr.03" }
func (adr03) Covers() []string { return []string{"AC-010-3"} }
func (adr03) Fix() string {
	return "在同一提交中新增或修改一条 ADR，并在其 Relates 中引用变化的模块或扩展点"
}

func (adr03) Check(s *workspace.Snapshot) []engine.Violation {
	sets, missing := changeSets(s)
	if missing != nil {
		return []engine.Violation{*missing}
	}
	var out []engine.Violation
	for _, cs := range sets {
		before, after := map[string]*model.Entity{}, map[string]*model.Entity{}
		refs := map[string]bool{}
		for _, f := range cs.Files {
			switch {
			case strings.HasPrefix(f.Path, "docs/architecture/") && strings.HasSuffix(f.Path, ".md"):
				collect(before, parseAt(f.Path, cs.Before(f.Path)))
				collect(after, parseAt(f.Path, cs.After(f.Path)))
			case strings.HasPrefix(f.Path, "docs/adr/") && f.Status != "D":
				if d := parseAt(f.Path, cs.After(f.Path)); d != nil {
					for _, e := range d.Entities {
						for _, r := range e.List("Relates") {
							refs[r] = true
						}
					}
				}
			}
		}
		for _, ch := range structuralChanges(before, after) {
			if refs[ch.id] {
				continue
			}
			file, line := "docs/architecture/overview.md", 1
			if e := after[ch.id]; e != nil {
				file, line = e.File, e.Line
			} else if e := before[ch.id]; e != nil {
				file = e.File
			}
			out = append(out, engine.Violation{File: file, Line: line,
				Msg: fmt.Sprintf("%s：%s %s，但同一提交中没有 ADR 引用它", cs.Name, ch.id, ch.what)})
		}
	}
	return out
}

func collect(m map[string]*model.Entity, d *model.Doc) {
	if d == nil {
		return
	}
	for _, e := range d.Entities {
		if e.Kind == "MOD" || e.Kind == "EXT" {
			m[e.ID] = e
		}
	}
}

type structural struct{ id, what string }

func structuralChanges(before, after map[string]*model.Entity) []structural {
	var out []structural
	ids := map[string]bool{}
	for id := range before {
		ids[id] = true
	}
	for id := range after {
		ids[id] = true
	}
	var sorted []string
	for id := range ids {
		sorted = append(sorted, id)
	}
	sort.Strings(sorted)
	for _, id := range sorted {
		b, a := before[id], after[id]
		switch {
		case b == nil:
			out = append(out, structural{id, "被新增"})
		case a == nil:
			out = append(out, structural{id, "被删除"})
		case a.Kind == "MOD" && setKey(b.List("Depends")) != setKey(a.List("Depends")):
			out = append(out, structural{id, fmt.Sprintf("的允许依赖由 [%s] 变为 [%s]", setKey(b.List("Depends")), setKey(a.List("Depends")))})
		case a.Kind == "EXT" && extKey(b) != extKey(a):
			out = append(out, structural{id, "的字段发生变化"})
		}
	}
	return out
}

func setKey(list []string) string {
	l := append([]string(nil), list...)
	sort.Strings(l)
	return strings.Join(l, ", ")
}

func extKey(e *model.Entity) string {
	var parts []string
	for _, f := range []string{"Status", "Interface", "Paths", "Closed", "Relates"} {
		parts = append(parts, f+"="+setKey(e.List(f)))
	}
	return strings.Join(parts, ";")
}
