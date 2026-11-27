package builtin

import (
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// task09：机械改动任务声明 Kind: mechanical、至少一条 Rewrite，并在 Relates 中引用说明原因的 ADR（AC-050-1）。
type task09 struct{}

func (task09) ID() string       { return "task.09" }
func (task09) Covers() []string { return []string{"AC-050-1"} }
func (task09) Fix() string {
	return "机械改动任务写 - Kind: mechanical、- Rewrite: <旧前缀> -> <新前缀>（多条用逗号分隔），并在 - Relates: 中引用说明改动原因的 ADR"
}

func (task09) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, t := range tasks(s) {
		kind, ok := t.Field("Kind")
		if !ok {
			continue
		}
		if strings.TrimSpace(kind) != "mechanical" {
			out = append(out, at(t, "%s 的 Kind 取值 %q 不合法，只能为 mechanical", t.ID, kind))
			continue
		}
		rs, bad := rewrites(t)
		for _, b := range bad {
			out = append(out, at(t, "%s 的 Rewrite %q 格式不对，应为 <旧前缀> -> <新前缀>", t.ID, b))
		}
		if len(rs) == 0 && len(bad) == 0 {
			out = append(out, at(t, "%s 是机械改动任务，但没有 Rewrite", t.ID))
		}
		adr := false
		for _, r := range t.List("Relates") {
			if e := s.Project.ByID[r]; e != nil && e.Kind == "ADR" {
				adr = true
			}
		}
		if !adr {
			out = append(out, at(t, "%s 是机械改动任务，但 Relates 中没有引用说明改动原因的 ADR", t.ID))
		}
	}
	return out
}
