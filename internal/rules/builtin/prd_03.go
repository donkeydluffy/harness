package builtin

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"harness/internal/lock"
	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// prd03：锁文件中的需求与用例编号不得从文档消失；Removed 的编号不得复用。
type prd03 struct{}

func (prd03) ID() string       { return "prd.03" }
func (prd03) Covers() []string { return []string{"AC-003-3", "AC-019-3"} }
func (prd03) Fix() string {
	return "不要删除或复用编号：需求改为 - Status: Removed，用例文字改为以 Removed 开头；新内容使用新编号"
}

var reqOrCaseRe = regexp.MustCompile(`^(FR|NFR)-\d+$|^AC-N?\d+-\d+/[NBE]\d+$`)

func (prd03) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	cur := s.LockItems()
	var ids []string
	for id := range s.Lock.Items {
		if reqOrCaseRe.MatchString(id) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		was := s.Lock.Items[id]
		now, ok := cur[id]
		if !ok {
			out = append(out, engine.Violation{File: workspace.PRDIndex, Line: 1,
				Msg: fmt.Sprintf("锁文件中的 %s 从文档中消失；编号只能标记 Removed，不能删除", id)})
			continue
		}
		if strings.HasPrefix(was, lock.RemovedPrefix) && !strings.HasPrefix(now, lock.RemovedPrefix) {
			file, line := locate(s, id)
			out = append(out, engine.Violation{File: file, Line: line,
				Msg: fmt.Sprintf("%s 已 Removed，编号不能复用", id)})
		}
	}
	// 文档中同一编号既有 Removed 又有新条目，也是复用。
	removed := map[string]bool{}
	for _, e := range requirements(s) {
		if workspace.Removed(e) {
			removed[e.ID] = true
		}
	}
	for _, e := range requirements(s) {
		if !workspace.Removed(e) && removed[e.ID] {
			out = append(out, at(e, "%s 已 Removed，编号不能复用", e.ID))
		}
	}
	return out
}

// locate 返回需求或用例所在的文件与行号，找不到时指向 PRD 索引。
func locate(s *workspace.Snapshot, id string) (string, int) {
	if c, ok := s.Project.Cases[id]; ok {
		return c.File, c.Line
	}
	if e, ok := s.Project.ByID[id]; ok {
		return e.File, e.Line
	}
	return workspace.PRDIndex, 1
}
