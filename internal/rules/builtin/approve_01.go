package builtin

import (
	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// approve01：状态为 Approved（ADR 为 Accepted）的产物必须有批准记录，且当前内容与批准哈希一致；
// 否则视为 Draft。
type approve01 struct{}

func (approve01) ID() string       { return "approve.01" }
func (approve01) Covers() []string { return []string{"AC-037-3"} }
func (approve01) Fix() string {
	return "由人在终端执行 harness approve <产物> 重新批准；不准备批准时把文件头 Status 改回 Draft（ADR 改回 Proposed）"
}

func (approve01) Check(s *workspace.Snapshot) []engine.Violation {
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
