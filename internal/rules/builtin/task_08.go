package builtin

import (
	"fmt"
	"regexp"
	"strings"

	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// task08：Done 任务的完成定义：所关联条目涉及的用例都有通过的集成测试；进度文件有该任务的记录
// （含结论与遗留问题）；除覆盖类规则外的全量检查通过。覆盖类规则 test.02、arch.05 要求全部用例与扩展点，
// 由实现阶段的退出规则保证，这里不计入，否则在全部任务完成前没有任务能 Done。
type task08 struct{}

func (task08) Needs() engine.Input { return engine.InputSources | engine.InputTests | engine.InputGit }
func (task08) ID() string          { return "task.08" }
func (task08) Covers() []string    { return []string{"AC-039-1", "AC-039-2", "AC-024-2"} }
func (task08) Fix() string {
	return "补齐失败或缺失的集成测试并重新生成测试报告；在 docs/progress.md 追加 \"### <日期> T-n\" 记录，含 - 结论： 与 - 遗留： 两行；修复全量检查的失败项。做不到时把任务改回 InProgress"
}

const progressFile = "docs/progress.md"

// 不计入完成定义中"全量检查"的规则。
var dodExcluded = []string{"test.02", "arch.05", "task.08"}

var progressHeadRe = regexp.MustCompile(`^### \d{4}-\d{2}-\d{2} (T-\d+)\b`)

func (task08) Check(s *workspace.Snapshot) []engine.Violation {
	var done []string
	var out []engine.Violation
	needTests := false
	for _, t := range tasks(s) {
		if t.Fields["Status"] == "Done" {
			done = append(done, t.ID)
			needTests = needTests || len(taskCases(s, t)) > 0
		}
	}
	if len(done) == 0 {
		return nil
	}
	if needTests && len(s.Reports) == 0 {
		return []engine.Violation{{File: ".", Line: 1, Msg: "输入缺失：没有测试报告（harness.json 的 testReports：" + strings.Join(s.Config.TestReports, ", ") + "）"}}
	}
	r := caseResults(s, staticTests(s))
	records := progressRecords(s)
	for _, t := range tasks(s) {
		if t.Fields["Status"] != "Done" {
			continue
		}
		var bad []string
		for _, id := range taskCases(s, t) {
			if r.violation(s, id) != nil {
				bad = append(bad, id)
			}
		}
		if len(bad) > 0 {
			out = append(out, at(t, "%s 为 Done，但这些用例没有通过的集成测试：%s", t.ID, strings.Join(bad, ", ")))
		}
		switch rec, ok := records[t.ID]; {
		case !ok:
			out = append(out, at(t, "%s 为 Done，但 %s 中没有该任务的记录", t.ID, progressFile))
		case !rec.conclusion || !rec.leftover:
			var miss []string
			if !rec.conclusion {
				miss = append(miss, "结论")
			}
			if !rec.leftover {
				miss = append(miss, "遗留")
			}
			out = append(out, engine.Violation{File: progressFile, Line: rec.line,
				Msg: fmt.Sprintf("%s 的进度记录缺少：%s", t.ID, strings.Join(miss, "、"))})
		}
	}
	var ids []string
	for _, id := range engine.IDs() {
		if !contains(dodExcluded, id) {
			ids = append(ids, id)
		}
	}
	if rep := engine.Run(s, engine.Selector{IDs: ids}); len(rep.Violations) > 0 {
		out = append(out, engine.Violation{File: tasksFile, Line: 1,
			Msg: fmt.Sprintf("全量检查（不含 %s）有 %d 条失败项，Done 的任务不满足完成定义：%s", strings.Join(dodExcluded, "、"), len(rep.Violations), strings.Join(done, ", "))})
	}
	return out
}

type progressRecord struct {
	line                 int
	conclusion, leftover bool
}

// progressRecords 返回每个任务的进度记录；同一任务有多条记录时，任意一条字段齐全即可。
func progressRecords(s *workspace.Snapshot) map[string]progressRecord {
	out := map[string]progressRecord{}
	d := s.Doc(progressFile)
	if d == nil {
		return out
	}
	var cur string
	var rec progressRecord
	flush := func() {
		if cur == "" {
			return
		}
		if old, ok := out[cur]; !ok || (!(old.conclusion && old.leftover) && rec.conclusion && rec.leftover) {
			out[cur] = rec
		}
	}
	for i, l := range d.Lines {
		if strings.HasPrefix(l, "#") {
			flush()
			cur = ""
			if m := progressHeadRe.FindStringSubmatch(l); m != nil {
				cur, rec = m[1], progressRecord{line: i + 1}
			}
			continue
		}
		rec.conclusion = rec.conclusion || strings.HasPrefix(l, "- 结论：")
		rec.leftover = rec.leftover || strings.HasPrefix(l, "- 遗留：")
	}
	flush()
	return out
}
