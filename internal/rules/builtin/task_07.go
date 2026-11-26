package builtin

import (
	"github.com/donkeydluffy/harness/internal/approval"
	"sort"
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// task07：InProgress 任务要求 PRD 有效批准、所引用设计元素所在产物有效批准、关联需求没有 Open 的待澄清项。
type task07 struct{}

func (task07) ID() string { return "task.07" }
func (task07) Covers() []string {
	return []string{"AC-038-1", "AC-038-2", "AC-038-3", "AC-011-5", "AC-047-4"}
}
func (task07) Fix() string {
	return "先由人用 harness approve 批准 PRD 与相关设计产物，并关闭关联需求的待澄清项，再把任务改为 InProgress；否则把任务改回 Todo"
}

func (task07) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, t := range tasks(s) {
		if t.Fields["Status"] != "InProgress" {
			continue
		}
		if s.Lock.ByUnit() {
			out = append(out, chainViolations(s, t.ID)...)
		} else if !effectivelyApproved(s, workspace.PRDIndex) {
			out = append(out, at(t, "%s 为 InProgress，但 PRD 没有处于有效的批准状态（Draft 或批准已失效）", t.ID))
		}
		files := map[string][]string{}
		for _, d := range t.List("Design") {
			if s.Lock.ByUnit() {
				break // 第 2 版按批准链判定，不再按文件
			}
			if e := s.Project.ByID[d]; e != nil && workspace.Approvable(e.File) && !effectivelyApproved(s, e.File) {
				files[e.File] = append(files[e.File], d)
			}
		}
		var names []string
		for f := range files {
			names = append(names, f)
		}
		sort.Strings(names)
		for _, f := range names {
			out = append(out, at(t, "%s 为 InProgress，但引用的 %s 所在的 %s 没有处于有效的批准状态", t.ID, strings.Join(files[f], ", "), f))
		}
		reqs := map[string]bool{}
		for _, c := range t.List("Covers") {
			if r := coveredRequirement(s, c); r != "" {
				reqs[r] = true
			}
		}
		for _, oq := range s.Project.ByKind["OQ"] {
			if oq.Fields["Status"] != "Open" {
				continue
			}
			for _, r := range oq.List("Relates") {
				if reqs[r] {
					out = append(out, at(t, "%s 为 InProgress，但关联需求 %s 有 Open 的待澄清项 %s", t.ID, r, oq.ID))
				}
			}
		}
	}
	return out
}

// chainViolations 在锁文件为第 2 版时要求 InProgress 任务自身与整条上游批准链有效（AC-047-4）。
func chainViolations(s *workspace.Snapshot, id string) []engine.Violation {
	var out []engine.Violation
	for _, st := range approval.Chain(s, id) {
		switch st.State {
		case approval.Unapproved:
			out = append(out, at(s.Project.ByID[id], "%s 为 InProgress，但批准链中的 %s 未批准", id, st.Unit.ID))
		case approval.Stale:
			out = append(out, at(s.Project.ByID[id], "%s 为 InProgress，但批准链中的 %s 批准已失效（%s 已变化）", id, st.Unit.ID, strings.Join(st.Cause, ", ")))
		}
	}
	return out
}
