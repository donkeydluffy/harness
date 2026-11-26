package builtin

import (
	"regexp"
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// task01：任务有编号、标题、状态、关联需求（FR、NFR 或 AC）、关联设计元素、文件范围、前置任务。
type task01 struct{}

func (task01) ID() string       { return "task.01" }
func (task01) Covers() []string { return []string{"AC-011-1"} }
func (task01) Fix() string {
	return "任务写作 \"### T-001 标题\"，字段 Status（Todo|InProgress|Done|Blocked）、Covers（FR、NFR 或 AC）、Design（设计元素）、Scope（文件范围 glob）、Depends（前置任务，可为空）；可选 Flow、Extends"
}

var taskIDRe = regexp.MustCompile(`^T-\d{3,}$`)

func (task01) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, t := range tasks(s) {
		if !taskIDRe.MatchString(t.ID) {
			out = append(out, at(t, "编号 %s 格式不合法，应为 T-001 形式", t.ID))
		}
		if strings.TrimSpace(t.Title) == "" {
			out = append(out, at(t, "%s 缺少标题", t.ID))
		}
		if st := t.Fields["Status"]; !oneOf(st, "Todo", "InProgress", "Done", "Blocked") {
			out = append(out, at(t, "%s 的 Status 取值 %q 不合法，合法取值：Todo、InProgress、Done、Blocked", t.ID, st))
		}
		for _, f := range []string{"Covers", "Design", "Scope"} {
			if len(t.List(f)) == 0 {
				out = append(out, at(t, "%s 缺少 %s", t.ID, f))
			}
		}
		if _, ok := t.Field("Depends"); !ok {
			out = append(out, at(t, "%s 缺少 Depends（没有前置任务时写 - Depends:）", t.ID))
		}
		for _, c := range t.List("Covers") {
			if msg := coverProblem(s, c); msg != "" {
				out = append(out, at(t, "%s 关联的 %s %s", t.ID, c, msg))
			}
		}
		for _, d := range t.List("Design") {
			if e := s.Project.ByID[d]; e == nil || !contains(designKinds, e.Kind) {
				out = append(out, at(t, "%s 关联的设计元素 %s 不存在", t.ID, d))
			}
		}
		refs := map[string][]string{"Depends": {"T"}, "Flow": {"FLOW"}, "Extends": {"EXT"}}
		for _, f := range []string{"Depends", "Flow", "Extends"} {
			for _, v := range t.List(f) {
				if e := s.Project.ByID[v]; e == nil || !contains(refs[f], e.Kind) {
					out = append(out, at(t, "%s 的 %s 引用的 %s 不存在", t.ID, f, v))
				}
			}
		}
	}
	out = append(out, ownDiags(s, "task.01")...)
	return out
}

// coverProblem 检查任务关联的需求编号，合法时返回空串。
func coverProblem(s *workspace.Snapshot, id string) string {
	if strings.HasPrefix(id, "AC-") {
		if it, req := acItem(s, id); it == nil {
			return "不存在"
		} else if workspace.Removed(req) { // Removed 标记只用于需求和用例，验收标准随需求删除
			return "已 Removed"
		}
		return ""
	}
	e := s.Project.ByID[id]
	switch {
	case e == nil || (e.Kind != "FR" && e.Kind != "NFR"):
		return "不存在（只能关联 FR、NFR 或 AC）"
	case workspace.Removed(e):
		return "已 Removed"
	}
	return ""
}
