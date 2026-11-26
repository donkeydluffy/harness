package builtin

import (
	"fmt"
	"sort"
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// debt03：Resolved 的技术债不能再被源码注释引用。
type debt03 struct{}

func (debt03) Needs() engine.Input { return engine.InputSources }
func (debt03) ID() string          { return "debt.03" }
func (debt03) Covers() []string    { return []string{"AC-031-3"} }
func (debt03) Fix() string {
	return "删除源码中对已解决技术债的引用，或把该技术债改回 Open"
}

func (debt03) Check(s *workspace.Snapshot) []engine.Violation {
	where := map[string][]string{}
	for _, c := range commentLines(s) {
		for _, e := range entityRefs(s, c.Text) {
			loc := fmt.Sprintf("%s:%d", c.File, c.Line)
			// 同一行用不同写法（如 DEBT-90 与 DEBT-090）引用同一项时只列一次位置
			if e.Kind == "DEBT" && e.Fields["Status"] == "Resolved" && !contains(where[e.ID], loc) {
				where[e.ID] = append(where[e.ID], loc)
			}
		}
	}
	var ids []string
	for id := range where {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var out []engine.Violation
	for _, id := range ids {
		out = append(out, at(s.Project.ByID[id], "%s 已 Resolved，仍被源码引用：%s", id, strings.Join(where[id], ", ")))
	}
	return out
}
