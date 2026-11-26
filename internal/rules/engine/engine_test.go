package engine

import (
	"strings"
	"testing"

	"github.com/donkeydluffy/harness/internal/model"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// exampleRule 演示 EXT-03：新增一条内置规则只需实现 Rule 并注册。
type exampleRule struct{ id, fix string }

func (r exampleRule) ID() string       { return r.id }
func (r exampleRule) Covers() []string { return []string{"AC-N009-1"} }
func (r exampleRule) Fix() string      { return r.fix }
func (r exampleRule) Check(s *workspace.Snapshot) []Violation {
	var out []Violation
	for _, e := range s.Project.ByKind["FR"] {
		if e.Fields["Priority"] == "" {
			out = append(out, Violation{File: e.File, Line: e.Line, Msg: e.ID + " 缺少 Priority"})
		}
	}
	return out
}

func snap(t *testing.T, src string) *workspace.Snapshot {
	t.Helper()
	d, _ := model.ParseMarkdown("docs/PRD.md", []byte(src))
	p, _ := model.Build([]*model.Doc{d})
	return &workspace.Snapshot{Project: p, Config: workspace.DefaultConfig()}
}

func TestExample_EXT_03(t *testing.T) {
	r := exampleRule{id: "example.01", fix: "补充 Priority 字段"}
	Register(r)
	if got, ok := Lookup("example.01"); !ok || got.ID() != "example.01" {
		t.Fatal("lookup failed")
	}
	bad := snap(t, "### FR-002 b\n- Status: Active\n\n### FR-001 a\n- Status: Active\n")
	rep := Run(bad, Selector{Prefixes: []string{"example"}})
	if len(rep.Ran) != 1 || len(rep.Violations) != 2 {
		t.Fatalf("report: %+v", rep)
	}
	if rep.Violations[0].Line != 1 || rep.Violations[0].Fix != "补充 Priority 字段" || rep.Violations[0].Rule != "example.01" {
		t.Fatalf("sorted/fix: %+v", rep.Violations)
	}
	ok := snap(t, "### FR-001 a\n- Priority: P0\n")
	if err := Verify(r, []Sample{{"bad", bad, true}, {"ok", ok, false}}); err != nil {
		t.Fatal(err)
	}
}

func TestRegisterDuplicatePanics_IF_011(t *testing.T) {
	Register(exampleRule{id: "example.02", fix: "x"})
	defer func() {
		if recover() == nil {
			t.Fatal("want panic")
		}
	}()
	Register(exampleRule{id: "example.02", fix: "x"})
}

func TestSelector_IF_012(t *testing.T) {
	sel := Selector{Prefixes: []string{"tasks"}}
	if !sel.match("task.01") || sel.match("tasx.01") {
		t.Fatal("plural prefix")
	}
	if !(Selector{}).match("any.01") {
		t.Fatal("empty selector matches all")
	}
	if u := (Selector{IDs: []string{"nope.01"}, Prefixes: []string{"nope"}}).Unknown(); len(u) != 2 {
		t.Fatalf("unknown: %v", u)
	}
}

type panicRule struct{ exampleRule }

func (panicRule) Check(*workspace.Snapshot) []Violation { panic("boom") }

func TestRunRecoversPanic_IF_012(t *testing.T) {
	Register(panicRule{exampleRule{id: "example.03", fix: "x"}})
	rep := Run(snap(t, ""), Selector{IDs: []string{"example.03"}})
	if len(rep.Errors) != 1 || !strings.Contains(rep.Errors[0].Error(), "boom") {
		t.Fatalf("errors: %v", rep.Errors)
	}
}

func TestVerifyRejects_IF_013(t *testing.T) {
	r := exampleRule{id: "example.04", fix: "x"}
	bad := snap(t, "### FR-001 a\n- Status: Active\n")
	if err := Verify(r, []Sample{{"bad", bad, true}}); err == nil {
		t.Fatal("want error: no ok sample")
	}
	if err := Verify(r, []Sample{{"bad", bad, false}, {"bad2", bad, true}}); err == nil {
		t.Fatal("want error: false positive")
	}
	if err := Verify(exampleRule{id: "example.05"}, []Sample{{"bad", bad, true}, {"ok", snap(t, ""), false}}); err == nil {
		t.Fatal("want error: empty fix")
	}
}

type warnRule struct{ exampleRule }

func (warnRule) Check(*workspace.Snapshot) []Violation {
	return []Violation{{File: "b.md", Line: 2, Msg: "告警", Warn: true}, {File: "a.md", Line: 1, Msg: "违规"}, {File: "a.md", Line: 3, Msg: "告警", Warn: true}}
}

// 告警放入 Warnings，不计入 Violations，各自排序。
func TestRunWarnings_IF_012(t *testing.T) {
	Register(warnRule{exampleRule{id: "example.06", fix: "x"}})
	rep := Run(snap(t, ""), Selector{IDs: []string{"example.06"}})
	if len(rep.Violations) != 1 || rep.Violations[0].Warn {
		t.Fatalf("violations: %+v", rep.Violations)
	}
	if len(rep.Warnings) != 2 || rep.Warnings[0].File != "a.md" || rep.Warnings[1].File != "b.md" {
		t.Fatalf("warnings: %+v", rep.Warnings)
	}
}
