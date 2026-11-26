package builtin

import (
	"fmt"
	"sort"
	"strings"

	"github.com/donkeydluffy/harness/internal/approval"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// approve01：状态为 Approved（ADR 为 Accepted）的产物必须有批准记录，且当前内容与批准哈希一致；
// 否则视为 Draft。
type approve01 struct{}

func (approve01) ID() string       { return "approve.01" }
func (approve01) Covers() []string { return []string{"AC-037-3", "AC-047-1", "AC-047-3"} }
func (approve01) Fix() string {
	return "由人在终端执行 harness approve <产物> 重新批准；不准备批准时把文件头 Status 改回 Draft（ADR 改回 Proposed）"
}

func (approve01) Check(s *workspace.Snapshot) []engine.Violation {
	if s.Lock.ByUnit() {
		return unitApprovals(s)
	}
	var out []engine.Violation
	for _, d := range s.Project.Docs {
		if !workspace.Approvable(d.Path) {
			continue
		}
		want := workspace.ApprovedStatus(d.Path)
		if workspace.HeaderStatus(d) != want {
			continue
		}
		line := workspace.HeaderStatusLine(d.Lines) + 1
		a := s.Lock.LastApproval(d.Path)
		switch {
		case a == nil:
			out = append(out, engine.Violation{File: d.Path, Line: line,
				Msg: "状态为 " + want + "，但锁文件中没有 harness approve 的批准记录，视为 Draft"})
		case a.Hash != s.ApprovalHash(d.Path):
			out = append(out, engine.Violation{File: d.Path, Line: line,
				Msg: "批准已失效：内容在 " + a.At.Format("2006-01-02 15:04") + " 由 " + a.By + " 批准后发生变化，视为 Draft"})
		}
	}
	return out
}

// unitApprovals 在锁文件为第 2 版时以告警列出未批准与批准失效的单位及原因（AC-047-1、AC-047-3）。
// 草稿随时可写（AC-047-5），所以只告警；任务能否开工由 task.07 判定。
func unitApprovals(s *workspace.Snapshot) []engine.Violation {
	states := approval.Evaluate(s)
	var ids []string
	for id := range states {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var out []engine.Violation
	for _, id := range ids {
		st := states[id]
		var msg string
		switch {
		case st.State == approval.Unapproved:
			msg = id + " 未批准"
		case st.State == approval.Stale && len(st.Cause) == 1 && st.Cause[0] == id:
			msg = id + " 批准已失效：内容在批准后发生变化"
		case st.State == approval.Stale:
			msg = fmt.Sprintf("%s 批准已失效：%s 已变化或未处于有效批准状态", id, strings.Join(without(st.Cause, id), ", "))
			if contains(st.Cause, id) {
				msg += "，自身内容也已变化"
			}
		default:
			continue
		}
		out = append(out, engine.Violation{File: st.Unit.File, Line: st.Unit.Line, Msg: msg, Warn: true,
			Fix: "由人在终端执行 harness approve " + id + "（先批准失效的上游）"})
	}
	return out
}

func without(list []string, x string) []string {
	var out []string
	for _, v := range list {
		if v != x {
			out = append(out, v)
		}
	}
	return out
}
