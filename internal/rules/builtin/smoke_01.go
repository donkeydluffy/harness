package builtin

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// smoke01：存在 harness.json 的项目已有 Done 任务时，项目配置须声明启动命令与冒烟测试命令；没有可启动应用的项目把启动命令写为
// "none: <原因>"，此时不要求冒烟测试命令（AC-055-1）。启动命令为 none 时任何时候都须写明原因。
type smoke01 struct{}

func (smoke01) ID() string       { return "smoke.01" }
func (smoke01) Covers() []string { return []string{"AC-055-1"} }
func (smoke01) Fix() string {
	return "由人在 harness.json 中声明 start（启动命令，或 \"none: <原因>\"）与 smoke（冒烟测试命令），并在 PRD 变更日志登记对象为 harness.json 的 CHG"
}

func (smoke01) Check(s *workspace.Snapshot) []engine.Violation {
	if _, err := os.Stat(filepath.Join(s.Root, workspace.ConfigFile)); err != nil {
		return nil // 与 ci.01、cfg.02 一致，只检查执行过 harness init 的项目
	}
	at := func(msg string) []engine.Violation {
		return []engine.Violation{{File: workspace.ConfigFile, Line: 1, Msg: msg}}
	}
	start := strings.TrimSpace(s.Config.Start)
	none := start == "none" || strings.HasPrefix(start, "none:")
	if none && strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(start, "none"), ":")) == "" {
		return at("启动命令声明为 none 但没有写明原因，应写作 \"none: <原因>\"")
	}
	var done string
	for _, t := range tasks(s) {
		if t.Fields["Status"] == "Done" {
			done = t.ID
			break
		}
	}
	switch {
	case done == "" || none:
		return nil
	case start == "":
		return at("已有 Done 的任务（" + done + "），但项目配置没有声明启动命令 start")
	case strings.TrimSpace(s.Config.Smoke) == "":
		return at("已有 Done 的任务（" + done + "），但项目配置没有声明冒烟测试命令 smoke")
	}
	return nil
}
