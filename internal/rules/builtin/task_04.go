package builtin

import (
	"strings"

	"harness/internal/model"
	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// task04：至多一个 InProgress 任务；两个未完成任务范围重叠时，须有一方（直接或间接）依赖另一方。
type task04 struct{}

func (task04) ID() string       { return "task.04" }
func (task04) Covers() []string { return []string{"AC-011-4"} }
func (task04) Fix() string {
	return "同一时间只保留一个 InProgress 任务；范围重叠的任务在 Depends 中声明先后，或拆开范围"
}

func (task04) Check(s *workspace.Snapshot) []engine.Violation {
	ts := tasks(s)
	byID := map[string]*model.Entity{}
	var active []string
	for _, t := range ts {
		byID[t.ID] = t
		if t.Fields["Status"] == "InProgress" {
			active = append(active, t.ID)
		}
	}
	var out []engine.Violation
	if len(active) > 1 {
		out = append(out, at(byID[active[1]], "同时有 %d 个 InProgress 任务：%s", len(active), strings.Join(active, ", ")))
	}
	for i, a := range ts {
		for _, b := range ts[i+1:] {
			if a.Fields["Status"] == "Done" || b.Fields["Status"] == "Done" {
				continue
			}
			var ov []string
			for _, p := range a.List("Scope") {
				for _, q := range b.List("Scope") {
					if globsOverlap(p, q) {
						ov = append(ov, p+" 与 "+q)
					}
				}
			}
			if len(ov) > 0 && !dependsOn(byID, a.ID, b.ID, nil) && !dependsOn(byID, b.ID, a.ID, nil) {
				out = append(out, at(b, "%s 与 %s 都未完成且范围重叠（%s），但彼此没有前置关系", a.ID, b.ID, strings.Join(ov, "；")))
			}
		}
	}
	return out
}

// dependsOn 判断任务 a 是否直接或间接依赖 b。
func dependsOn(byID map[string]*model.Entity, a, b string, seen map[string]bool) bool {
	if seen == nil {
		seen = map[string]bool{}
	}
	if seen[a] || byID[a] == nil {
		return false
	}
	seen[a] = true
	for _, d := range byID[a].List("Depends") {
		if d == b || dependsOn(byID, d, b, seen) {
			return true
		}
	}
	return false
}
