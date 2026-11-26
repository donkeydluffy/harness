package builtin

import (
	"strconv"

	"github.com/donkeydluffy/harness/internal/model"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// prd12：需求取代关系（AC-003-6）。Replaces 引用存在的需求或验收标准，被取代者的 Release 小于取代者；
// 被取代的需求用 Replaced-By 指回取代者，两边互相对应。
type prd12 struct{}

func (prd12) ID() string       { return "prd.12" }
func (prd12) Covers() []string { return []string{"AC-003-6"} }
func (prd12) Fix() string {
	return "在被取代的需求上写 - Replaced-By: <取代者>，在取代者的 Replaces 中列出它；被取代者的 Release 须小于取代者的 Release"
}

func (prd12) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, e := range requirements(s) {
		for _, id := range e.List("Replaces") {
			owner := replacedOwner(s, id)
			switch {
			case owner == nil:
				out = append(out, at(e, "%s 的 Replaces 引用了不存在的 %s", e.ID, id))
				continue
			case release(owner) >= release(e):
				out = append(out, at(e, "%s 取代 %s，但被取代者的 Release（%d）不小于取代者的 Release（%d）", e.ID, id, release(owner), release(e)))
			}
			if owner.ID == id && !contains(owner.List("Replaced-By"), e.ID) {
				out = append(out, at(owner, "%s 被 %s 取代，但缺少 Replaced-By: %s", id, e.ID, e.ID))
			}
		}
		for _, by := range e.List("Replaced-By") {
			r := s.Project.ByID[by]
			if r == nil || !contains(r.List("Replaces"), e.ID) {
				out = append(out, at(e, "%s 的 Replaced-By 为 %s，但 %s 的 Replaces 不含 %s", e.ID, by, by, e.ID))
			}
		}
	}
	return out
}

// replacedOwner 返回被取代条目所在的需求：id 是需求时为它本身，是验收标准时为其所属需求；不存在时为 nil。
func replacedOwner(s *workspace.Snapshot, id string) *model.Entity {
	for _, e := range requirements(s) {
		if e.ID == id {
			return e
		}
		for _, it := range e.Items {
			if it.ID == id {
				return e
			}
		}
	}
	return nil
}

func release(e *model.Entity) int {
	n, _ := strconv.Atoi(e.Fields["Release"])
	return n
}
