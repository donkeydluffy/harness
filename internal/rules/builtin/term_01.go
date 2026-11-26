package builtin

import (
	"fmt"
	"strings"

	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// term01：受检产物（PRD、架构声明、详细设计、任务列表）中不得出现术语表的禁用别名。
// 代码格式中的文字与统一语言表格本身不检查。
type term01 struct{}

func (term01) ID() string       { return "term.01" }
func (term01) Covers() []string { return []string{"AC-005-1"} }
func (term01) Fix() string {
	return "改用统一语言表格中\"术语\"列的名称；确需字面引用时放进代码格式"
}

// termChecked 判断产物是否受统一语言约束。
func termChecked(rel string) bool {
	return workspace.IsPRDPath(rel) || strings.HasPrefix(rel, "docs/architecture/") || rel == "docs/tasks.md"
}

func (term01) Check(s *workspace.Snapshot) []engine.Violation {
	banned := map[string]string{}
	var words []string
	for _, t := range s.Project.Terms {
		for _, b := range t.Banned {
			if _, ok := banned[b]; !ok {
				words = append(words, b)
			}
			banned[b] = t.Name
		}
	}
	if len(words) == 0 {
		return nil
	}
	var out []engine.Violation
	for _, d := range s.Project.Docs {
		if !termChecked(d.Path) {
			continue
		}
		lines, nums := proseLines(d.Lines, 1)
		for i, l := range lines {
			if strings.HasPrefix(strings.TrimSpace(l), "| TERM") {
				continue
			}
			for _, w := range words {
				if strings.Contains(l, w) {
					out = append(out, engine.Violation{File: d.Path, Line: nums[i],
						Msg: fmt.Sprintf("出现禁用别名 %q，应使用术语 %q", w, banned[w])})
				}
			}
		}
	}
	return out
}
