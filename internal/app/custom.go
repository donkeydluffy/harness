package app

import (
	"sort"
	"strings"

	"github.com/donkeydluffy/harness/internal/approval"
	"github.com/donkeydluffy/harness/internal/rules/custom"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// projectRules 构造项目自定义规则（IF-036）并用样本验证（IF-013）。返回可运行的规则，以及定义有误、
// 样本不符合预期、与锁文件不一致的问题（AC-054-1 ~ AC-054-4）；有问题的规则不运行。
func projectRules(s *workspace.Snapshot) ([]engine.Rule, []engine.Violation) {
	rules, bad := custom.Load(s, engine.IDs())
	verified := custom.Verify(s.Root, rules)
	failed := map[string]bool{}
	for _, v := range verified {
		failed[v.Rule] = true
	}
	bad = append(bad, verified...)
	var ok []engine.Rule
	for _, r := range rules {
		if !failed[r.ID()] {
			ok = append(ok, r)
		}
	}
	if len(s.Lock.Items) > 0 { // 空基线：首次 harness lock 之前不追踪
		for key, h := range s.LockItems() {
			id, isRule := strings.CutPrefix(key, "rule:")
			if !isRule || s.Lock.Items[key] == h {
				continue
			}
			what := "的定义或样本与锁文件不一致"
			if s.Lock.Items[key] == "" {
				what = "是锁定之后新增的"
			}
			bad = append(bad, engine.Violation{Rule: id, File: workspace.ConfigFile, Line: 1,
				Msg: "自定义规则 " + id + what, Fix: "确认规则定义与样本后运行 harness lock"})
		}
	}
	return ok, bad
}

// runProject 把选中的项目规则及其问题加入报告（AC-054-2）。
func runProject(s *workspace.Snapshot, sel engine.Selector, rules []engine.Rule, bad []engine.Violation, rep *engine.Report) {
	for _, v := range bad {
		if selected(sel, v.Rule) {
			rep.Violations = append(rep.Violations, v)
		}
	}
	for _, r := range rules {
		if !selected(sel, r.ID()) {
			continue
		}
		rep.Ran = append(rep.Ran, r.ID())
		rep.Violations = append(rep.Violations, check(r, s)...)
	}
	sort.SliceStable(rep.Violations, func(i, j int) bool {
		a, b := rep.Violations[i], rep.Violations[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Rule < b.Rule
	})
}

// check 运行一条项目规则，补齐规则编号与修复建议。
func check(r engine.Rule, s *workspace.Snapshot) []engine.Violation {
	vs := r.Check(s)
	for i := range vs {
		vs[i].Rule = r.ID()
		if vs[i].Fix == "" {
			vs[i].Fix = r.Fix()
		}
	}
	return vs
}

// selected 判断选择器是否选中规则编号，与 engine 的选择规则一致：空选择器选中全部，前缀可用复数写法。
func selected(sel engine.Selector, id string) bool {
	if len(sel.IDs) == 0 && len(sel.Prefixes) == 0 {
		return true
	}
	for _, x := range sel.IDs {
		if x == id {
			return true
		}
	}
	for _, p := range sel.Prefixes {
		if strings.HasPrefix(id, p+".") || (strings.HasSuffix(p, "s") && strings.HasPrefix(id, strings.TrimSuffix(p, "s")+".")) {
			return true
		}
	}
	return false
}

// approveChecks 运行登记在待批准单位类型上的项目规则，只保留落在这些单位内的违规（AC-054-5）。
func approveChecks(s *workspace.Snapshot, pending []approval.Pending) []engine.Violation {
	rules, _ := projectRules(s)
	var out []engine.Violation
	for _, r := range rules {
		vs := check(r, s)
		for _, p := range pending {
			if len(custom.ApproveChecks([]engine.Rule{r}, string(p.Unit.Kind))) == 0 {
				continue
			}
			for _, v := range vs {
				if p.Unit.Owns(v.File, v.Line) {
					out = append(out, v)
				}
			}
		}
	}
	return out
}
