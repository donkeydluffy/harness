package builtin

import (
	"sort"
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// task05：任务范围只落在一个模块内；跨模块时须引用一个包含全部涉及模块的调用流程。
type task05 struct{}

func (task05) ID() string       { return "task.05" }
func (task05) Covers() []string { return []string{"AC-011-6", "AC-050-2"} }
func (task05) Fix() string {
	return "把任务拆成每个模块一个，或在 Flow 中引用一个包含全部涉及模块的调用流程"
}

func (task05) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, t := range tasks(s) {
		if mechanical(t) {
			continue // 机械改动任务免于调用流程限制（AC-050-2）
		}
		touched := map[string]bool{}
		for _, p := range t.List("Scope") {
			for _, m := range scopeOwners(s, p) {
				touched[m] = true
			}
		}
		if len(touched) <= 1 {
			continue
		}
		var mods []string
		for m := range touched {
			mods = append(mods, m)
		}
		sort.Strings(mods)
		flows := taskFlows(t)
		if len(flows) == 0 {
			out = append(out, at(t, "%s 的文件范围跨 %d 个模块（%s），但没有引用调用流程", t.ID, len(mods), strings.Join(mods, ", ")))
			continue
		}
		var best []string
		for _, f := range flows {
			fm := flowModules(s, f)
			var missing []string
			for _, m := range mods {
				if !fm[m] {
					missing = append(missing, m)
				}
			}
			if len(missing) == 0 {
				best = nil
				break
			}
			if best == nil || len(missing) < len(best) {
				best = append([]string{f}, missing...)
			}
		}
		if best != nil {
			out = append(out, at(t, "%s 涉及的模块 %s 不在引用的流程 %s 中", t.ID, strings.Join(best[1:], ", "), best[0]))
		}
	}
	return out
}
