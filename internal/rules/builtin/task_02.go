package builtin

import (
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// task02：纳入范围的 Active FR 被任务覆盖：直接关联，或其每条有效验收标准都被关联（ADR-008）。
// 还没有任务列表时不检查。
type task02 struct{}

func (task02) ID() string       { return "task.02" }
func (task02) Covers() []string { return []string{"AC-011-2"} }
func (task02) Fix() string {
	return "在某个任务的 Covers 中关联该 FR，或关联它尚未被关联的验收标准"
}

func (task02) Check(s *workspace.Snapshot) []engine.Violation {
	if s.Doc(tasksFile) == nil {
		return nil
	}
	covered := map[string]bool{}
	for _, t := range tasks(s) {
		for _, c := range t.List("Covers") {
			covered[c] = true
		}
	}
	var out []engine.Violation
	for _, e := range requirements(s) {
		if e.Kind != "FR" || e.Fields["Status"] != "Active" || !inScope(s, e) || replacedBy(s, e.ID) != "" || covered[e.ID] {
			continue
		}
		var missing []string
		for _, it := range e.Items {
			if strings.HasPrefix(it.ID, "AC-") && !covered[it.ID] && replacedBy(s, it.ID) == "" {
				missing = append(missing, it.ID)
			}
		}
		if len(missing) > 0 || len(e.Items) == 0 {
			out = append(out, at(e, "%s 没有被任务覆盖，未被关联的验收标准：%s", e.ID, strings.Join(missing, ", ")))
		}
	}
	return out
}
