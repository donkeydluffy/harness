package app

import (
	"fmt"
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/toolrun"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// MissingTool 是工具未安装时 Run 返回的错误，供 cli 给出安装提示（AC-014-2）。
type MissingTool = toolrun.MissingTool

// ToolKinds 是会运行外部工具的检查（IF-038）。
var ToolKinds = []string{"lint", "unit", "smoke"}

// Run 运行外部工具并写报告（kind 为 lint、unit、smoke，调用 IF-030 ~ IF-032），
// 然后重新加载快照、运行同名区域的规则并返回结果（IF-038）。工具未安装时返回 *toolrun.MissingTool。
//
// harness:IF-038
func Run(root, kind string) (engine.Report, error) {
	registerRules()
	opt := workspace.LoadOptions{WithSources: true, WithTests: true}
	s, err := workspace.Load(root, opt)
	if err != nil {
		return engine.Report{}, err
	}
	var extra []engine.Violation
	switch kind {
	case "lint":
		_, err = toolrun.Lint(s)
	case "unit":
		_, err = toolrun.Unit(s)
	case "smoke":
		if strings.TrimSpace(s.Config.Smoke) == "" {
			break // 未声明冒烟测试命令：是否必须声明由 smoke.01 判定（AC-055-1）
		}
		var code int
		var tail []string
		code, tail, err = toolrun.Smoke(s)
		if err == nil && code != 0 {
			extra = append(extra, engine.Violation{Rule: "smoke", File: workspace.ConfigFile, Line: 1,
				Msg: fmt.Sprintf("冒烟测试命令 %s 失败，退出码 %d，输出的最后几行：\n%s", s.Config.Smoke, code, strings.Join(tail, "\n")),
				Fix: "修复启动或冒烟测试失败的原因后重新运行 harness check smoke"})
		}
	default:
		return engine.Report{}, fmt.Errorf("不支持的检查 %s（可用：%s）", kind, strings.Join(ToolKinds, ", "))
	}
	if err != nil {
		return engine.Report{}, err
	}
	if s, err = workspace.Load(root, opt); err != nil {
		return engine.Report{}, err
	}
	rep := engine.Run(s, engine.Selector{Prefixes: []string{kind}})
	if len(rep.Errors) > 0 {
		return rep, rep.Errors[0]
	}
	rep.Violations = append(rep.Violations, extra...)
	return rep, nil
}
