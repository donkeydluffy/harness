package app

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/donkeydluffy/harness/internal/approval"
	"github.com/donkeydluffy/harness/internal/lock"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// ErrNothingPending 表示范围内没有待批准的单位（AC-047-2/B1）。
var ErrNothingPending = errors.New("没有待批准的内容")

// downstreamRules 要求下游已覆盖上游，批准单位时不运行（AC-047-2）。
var downstreamRules = []string{"trace.01", "task.02"}

// approveUnits 按批准单位批准（锁文件第 2 版，ADR-009）：列出 target 范围内待批准的单位与差异，
// 运行这些单位自身的检查并要求上游已批准，确认后写入批准记录。不修改产物文件。
func approveUnits(root string, s *workspace.Snapshot, target string, t Terminal) error {
	pending, err := approval.PendingIn(s, target)
	if err != nil {
		return err
	}
	if len(pending) == 0 {
		return ErrNothingPending
	}
	if rep := unitViolations(s, pending); len(rep.Violations) > 0 {
		return &ViolationsError{Report: rep}
	}
	states := approval.Evaluate(s)
	var b strings.Builder
	units := make([]approval.Unit, len(pending))
	for i, p := range pending {
		units[i] = p.Unit
		st := states[p.Unit.ID]
		switch {
		case p.Before == "":
			fmt.Fprintf(&b, "新增 %s（%s:%d）\n", p.Unit.ID, p.Unit.File, p.Unit.Line)
		case p.Before != p.Unit.Text:
			fmt.Fprintf(&b, "修改 %s（%s:%d）\n", p.Unit.ID, p.Unit.File, p.Unit.Line)
		default:
			fmt.Fprintf(&b, "失效 %s（%s:%d）：上游 %s 已变化\n", p.Unit.ID, p.Unit.File, p.Unit.Line, strings.Join(st.Cause, ", "))
		}
		for _, l := range lineDiff(p.Before, p.Unit.Text) {
			b.WriteString("    " + l + "\n")
		}
	}
	fmt.Fprintf(&b, "批准以上 %d 个单位？输入 y 确认：", len(units))
	if !t.Confirm(b.String()) {
		return ErrDeclined
	}
	if s.Lock.Units == nil {
		s.Lock.Units = map[string]lock.UnitApproval{}
	}
	for id, rec := range approval.Record(s, units, t.Identity(), now()) {
		s.Lock.Units[id] = rec
	}
	return s.Lock.Write(filepath.Join(root, lock.File))
}

// unitViolations 返回待批准单位自身的检查结果：产物文档规则中不要求下游覆盖的那些，只保留落在这些单位内的违规；
// 再加上未批准的上游（AC-047-3）。同一批中一起批准的上游视为已批准。
func unitViolations(s *workspace.Snapshot, pending []approval.Pending) engine.Report {
	skip := map[string]bool{}
	for _, id := range append(append([]string{}, downstreamRules...), engine.CoverageRules...) {
		skip[id] = true
	}
	var ids []string
	for _, id := range docRules().IDs {
		if !skip[id] && !strings.HasPrefix(id, "approve.") {
			ids = append(ids, id)
		}
	}
	rep := engine.Run(s, engine.Selector{IDs: ids})
	var keep []engine.Violation
	for _, v := range rep.Violations {
		for _, p := range pending {
			if p.Unit.Owns(v.File, v.Line) {
				keep = append(keep, v)
				break
			}
		}
	}
	batch := map[string]bool{}
	for _, p := range pending {
		batch[p.Unit.ID] = true
	}
	states := approval.Evaluate(s)
	for _, p := range pending {
		for _, up := range p.Unit.Upstream {
			if st := states[up]; !batch[up] && st.State != approval.Approved {
				keep = append(keep, engine.Violation{Rule: "approve.01", File: p.Unit.File, Line: p.Unit.Line,
					Msg: fmt.Sprintf("%s 引用的 %s %s，不能批准", p.Unit.ID, up, describe(st)),
					Fix: "先批准上游：harness approve " + up})
			}
		}
	}
	rep.Violations = keep
	return rep
}

// describe 用文字说明非 approved 的状态。
func describe(st approval.Status) string {
	if st.State == approval.Unapproved {
		return "未批准"
	}
	return "批准已失效（" + strings.Join(st.Cause, ", ") + " 已变化）"
}

// lineDiff 返回 before 到 after 的逐行差异，删除行前加 "- "，新增行前加 "+ "，相同行省略。
func lineDiff(before, after string) []string {
	a, b := split(before), split(after)
	// 最长公共子序列
	n, m := len(a), len(b)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	var out []string
	i, j := 0, 0
	for i < n || j < m {
		switch {
		case i < n && j < m && a[i] == b[j]:
			i, j = i+1, j+1
		case j < m && (i == n || lcs[i][j+1] >= lcs[i+1][j]):
			out = append(out, "+ "+b[j])
			j++
		default:
			out = append(out, "- "+a[i])
			i++
		}
	}
	return out
}

func split(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}
