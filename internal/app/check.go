package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// CheckOptions 是 harness check 的参数（IF-022）。
type CheckOptions struct {
	Rules       []string // 规则编号（含点，如 prd.01）或前缀（如 prd、tasks）；空表示全部
	Commits     string   // 提交范围，如 origin/main..HEAD；为空时检查暂存区
	MessageFile string   // 即将形成的提交的信息文件（commit-msg hook 的参数）
	JSON        bool
}

// Check 加载项目快照并运行选中的规则（IF-022）。
// 返回的 error 表示内部错误（如产物不可读、选择器无效），与"有违规"区分。
func Check(root string, opt CheckOptions) (engine.Report, error) {
	registerRules()
	sel := selector(opt.Rules)
	lo := workspace.LoadOptions{CommitRange: opt.Commits, WithSources: true, WithTests: true}
	if opt.MessageFile != "" {
		b, err := os.ReadFile(opt.MessageFile)
		if err != nil {
			return engine.Report{}, fmt.Errorf("读取提交信息：%w", err)
		}
		lo.Message = stripComments(string(b))
	}
	s, err := workspace.Load(root, lo)
	if err != nil {
		return engine.Report{}, err
	}
	rules, bad := projectRules(s)
	var unknown []string
	for _, u := range sel.Unknown() {
		if !projectSelected(u, rules, bad) {
			unknown = append(unknown, u)
		}
	}
	if len(unknown) > 0 {
		return engine.Report{}, fmt.Errorf("没有匹配的规则：%s（可用规则：%s）",
			strings.Join(unknown, ", "), strings.Join(engine.IDs(), ", "))
	}
	rep := engine.Run(s, sel)
	if len(rep.Errors) > 0 {
		return rep, rep.Errors[0]
	}
	runProject(s, sel, rules, bad, &rep)
	return rep, nil
}

// projectSelected 判断选择器条目是否匹配某条项目规则（含定义有误的规则）。
func projectSelected(item string, rules []engine.Rule, bad []engine.Violation) bool {
	one := selector([]string{item})
	for _, r := range rules {
		if selected(one, r.ID()) {
			return true
		}
	}
	for _, v := range bad {
		if selected(one, v.Rule) {
			return true
		}
	}
	return false
}

// stripComments 去掉提交信息中 git 的注释行（以 # 开头）。
func stripComments(msg string) string {
	var keep []string
	for _, l := range strings.Split(msg, "\n") {
		if !strings.HasPrefix(l, "#") {
			keep = append(keep, l)
		}
	}
	return strings.TrimSpace(strings.Join(keep, "\n"))
}

// docRules 选中只读取产物文档的规则（不需要源码、测试报告、git）。
func docRules() engine.Selector {
	var sel engine.Selector
	for _, id := range engine.IDs() {
		if r, _ := engine.Lookup(id); engine.Needs(r) == 0 {
			sel.IDs = append(sel.IDs, id)
		}
	}
	return sel
}

func run(root string, sel engine.Selector, opt workspace.LoadOptions) (engine.Report, error) {
	s, err := workspace.Load(root, opt)
	if err != nil {
		return engine.Report{}, err
	}
	rep := engine.Run(s, sel)
	if len(rep.Errors) > 0 {
		return rep, rep.Errors[0]
	}
	return rep, nil
}

func selector(args []string) engine.Selector {
	var sel engine.Selector
	for _, a := range args {
		if strings.Contains(a, ".") {
			sel.IDs = append(sel.IDs, a)
		} else {
			sel.Prefixes = append(sel.Prefixes, a)
		}
	}
	return sel
}
