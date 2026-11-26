package builtin

import (
	"sort"
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// if04：接口快照中的签名相对锁文件变化时，须有锁定之后新增的、引用该接口的 CHG 或 ADR；
// 删除接口或修改签名的 CHG 说明以 Breaking: 开头。违规信息列出依赖该接口所属模块的下游模块（AC-052-4）。
type if04 struct{}

func (if04) ID() string       { return "if.04" }
func (if04) Covers() []string { return []string{"AC-052-4"} }
func (if04) Fix() string {
	return "在 PRD 变更日志新增引用该接口的 CHG（删除接口或修改签名时说明以 Breaking: 开头），或新增 Relates 含该接口的 ADR；确认后运行 harness lock"
}

const snapshotPrefix = "snapshot:"

func (if04) Check(s *workspace.Snapshot) []engine.Violation {
	base := s.Lock.Items
	tracked := false
	for k := range base {
		if strings.HasPrefix(k, snapshotPrefix) {
			tracked = true
			break
		}
	}
	if !tracked {
		return nil // 空基线：锁文件还没有记录接口快照
	}
	now := s.LockItems()
	ids := map[string]bool{}
	for _, items := range []map[string]string{base, now} {
		for k := range items {
			if strings.HasPrefix(k, snapshotPrefix) {
				ids[strings.TrimPrefix(k, snapshotPrefix)] = true
			}
		}
	}
	var sorted []string
	for id := range ids {
		sorted = append(sorted, id)
	}
	sort.Strings(sorted)
	var out []engine.Violation
	for _, id := range sorted {
		was, is := base[snapshotPrefix+id], now[snapshotPrefix+id]
		if was == is {
			continue
		}
		what, breaking := "签名有变化", true
		switch {
		case was == "":
			what, breaking = "是新增的接口", false
		case is == "":
			what = "已删除"
		}
		var chg, marked, adr bool
		for _, c := range s.Project.Changes {
			if _, old := base[c.ID]; !old && contains(c.Targets, id) {
				chg = true
				marked = marked || strings.HasPrefix(strings.TrimSpace(c.Note), "Breaking:")
			}
		}
		for _, e := range s.Project.ByKind["ADR"] {
			if _, old := base[e.ID]; !old && e.Fields["Status"] != "Superseded" && contains(e.List("Relates"), id) {
				adr = true
			}
		}
		msg := ""
		switch {
		case chg && breaking && !marked:
			msg = id + " " + what + "，登记它的 CHG 说明没有以 Breaking: 开头"
		case chg || adr:
			continue
		default:
			msg = id + " " + what + "，但没有引用它的新 CHG 条目或 ADR"
		}
		if down := downstream(s, id); len(down) > 0 {
			msg += "；受影响的下游模块：" + strings.Join(down, ", ")
		}
		out = append(out, engine.Violation{File: workspace.InterfacesFile, Line: snapshotLine(s, id), Msg: msg})
	}
	return out
}

// downstream 返回 Depends 中含接口所属模块的模块，按编号排序。
func downstream(s *workspace.Snapshot, ifID string) []string {
	e := s.Project.ByID[ifID]
	if e == nil {
		return nil
	}
	mod := e.Fields["Module"]
	var out []string
	for _, m := range archEntities(s, "MOD") {
		if effective(m) && contains(m.List("Depends"), mod) {
			out = append(out, m.ID)
		}
	}
	sort.Strings(out)
	return out
}
