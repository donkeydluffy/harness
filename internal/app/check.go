package app

import (
	"fmt"
	"strings"

	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// CheckOptions 是 harness check 的参数（IF-022）。
type CheckOptions struct {
	Rules   []string // 规则编号（含点，如 prd.01）或前缀（如 prd、tasks）；空表示全部
	Commits string
	JSON    bool
}

// Check 加载项目快照并运行选中的规则（IF-022）。
// 返回的 error 表示内部错误（如产物不可读、选择器无效），与"有违规"区分。
func Check(root string, opt CheckOptions) (engine.Report, error) {
	registerRules()
	sel := selector(opt.Rules)
	if unknown := sel.Unknown(); len(unknown) > 0 {
		return engine.Report{}, fmt.Errorf("没有匹配的规则：%s（可用规则：%s）",
			strings.Join(unknown, ", "), strings.Join(engine.IDs(), ", "))
	}
	return run(root, sel, workspace.LoadOptions{CommitRange: opt.Commits, WithSources: true, WithTests: true})
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
