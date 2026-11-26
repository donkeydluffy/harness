package builtin

import (
	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// cfg01：harness.json 的规范化哈希相对锁文件变化时，须有锁定之后新增的 CHG 条目（对象写 harness.json）
// 或新增的 ADR（Relates 含 harness.json）说明这次变化。
type cfg01 struct{}

func (cfg01) ID() string       { return "cfg.01" }
func (cfg01) Covers() []string { return []string{"AC-041-1"} }
func (cfg01) Fix() string {
	return "在 PRD 变更日志中新增一行 CHG，对象写 harness.json，说明改了哪项门禁配置及原因（或新增一条 Relates 含 harness.json 的 ADR）；确认后运行 harness lock"
}

func (cfg01) Check(s *workspace.Snapshot) []engine.Violation {
	base := s.Lock.Items
	if len(base) == 0 {
		return nil // 空基线：首次 harness lock 之前不追踪
	}
	was, now := base["config"], s.LockItems()["config"]
	if was == now {
		return nil
	}
	for _, c := range s.Project.Changes {
		if _, old := base[c.ID]; !old && contains(c.Targets, workspace.ConfigFile) {
			return nil
		}
	}
	for _, e := range s.Project.ByKind["ADR"] {
		if _, old := base[e.ID]; !old && e.Fields["Status"] != "Superseded" && contains(e.List("Relates"), workspace.ConfigFile) {
			return nil
		}
	}
	what := "内容有变化"
	switch {
	case was == "":
		what = "是锁定之后新增的"
	case now == "":
		what = "被删除"
	}
	return []engine.Violation{{File: workspace.ConfigFile, Line: 1,
		Msg: "门禁配置 " + workspace.ConfigFile + " 相对锁文件" + what + "，但没有引用它的新 CHG 条目或 ADR"}}
}
