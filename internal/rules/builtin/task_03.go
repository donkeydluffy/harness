package builtin

import (
	"sort"
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// task03：任务的文件范围落在其关联模块的路径内。
type task03 struct{}

func (task03) ID() string       { return "task.03" }
func (task03) Covers() []string { return []string{"AC-050-2", "AC-007-5"} }
func (task03) Fix() string {
	return "把范围收窄到关联模块的 Paths 内，或在 Design 中关联该路径所属的模块、接口或流程"
}

func (task03) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, t := range tasks(s) {
		if mechanical(t) || t.Fields["Status"] == "Done" {
			continue // 机械改动任务免于模块范围限制（AC-050-2）；Done 的任务是历史记录（AC-007-5）
		}
		assoc := taskModules(s, t)
		for _, p := range t.List("Scope") {
			ok := false
			for _, m := range scopeOwners(s, p) {
				ok = ok || assoc[m]
			}
			if !ok {
				var mods []string
				for m := range assoc {
					mods = append(mods, m)
				}
				sort.Strings(mods)
				out = append(out, at(t, "%s 的文件范围 %s 不在关联模块（%s）的路径内", t.ID, p, strings.Join(mods, ", ")))
			}
		}
	}
	return out
}
