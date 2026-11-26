package app

import (
	"fmt"
	"strings"

	"github.com/donkeydluffy/harness/internal/gate"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// Gate 加载轻量快照并调用门禁判定（IF-025）；cli 负责用 adapter 解码代理输入、编码判定结果。
// Bash 命令放行前记录命令前快照，命令执行后比对并恢复不被允许的改动（FLOW-03）。
// 任何内部错误都判定为拒绝（失败即阻断）。
func Gate(root string, req gate.Request) gate.Decision {
	registerRules()
	// 会话开始与结束运行全量检查，需要源码与测试报告；其余请求用轻量快照。
	full := req.Kind == gate.Stop || req.Kind == gate.SessionStart
	s, err := workspace.Load(root, workspace.LoadOptions{WithSources: full, WithTests: full})
	if err != nil {
		return gate.Decision{Rule: "gate.error", Reason: "无法加载项目：" + err.Error()}
	}
	switch req.Kind {
	case gate.Command:
		d := gate.Decide(s, req)
		if d.Allow {
			if err := gate.Before(root, req.SessionID); err != nil {
				return gate.Decision{Rule: "gate.error", Reason: "无法记录命令前快照：" + err.Error()}
			}
		}
		return d
	case gate.PostWrite:
		if req.Command == "" {
			return gate.Decide(s, req)
		}
		changes, err := gate.After(root, req.SessionID, gate.AfterAllowed(s, req.Command))
		if err != nil {
			return gate.Decision{Rule: "gate.error", Reason: "命令后比对失败：" + err.Error()}
		}
		return gate.Decision{Allow: true, Feedback: restoreReport(changes)}
	default:
		return gate.Decide(s, req)
	}
}

// restoreReport 列出命令改动中被恢复或无法恢复的文件；全部被允许时为空。
func restoreReport(changes []gate.Change) string {
	var lines []string
	for _, c := range changes {
		switch {
		case c.Restored:
			lines = append(lines, fmt.Sprintf("- %s（%s）已恢复为命令前状态", c.Path, c.Action))
		case c.Reason != "允许":
			lines = append(lines, fmt.Sprintf("- %s（%s）：%s", c.Path, c.Action, c.Reason))
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return "[gate.write] 命令改动了不允许写入的文件：\n" + strings.Join(lines, "\n")
}
