package builtin

import (
	"sort"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// cfg03：项目配置 tools 中的语言都在 languages 中，且只覆盖 format、lint（AC-014-3）。
type cfg03 struct{}

func (cfg03) ID() string       { return "cfg.03" }
func (cfg03) Covers() []string { return []string{"AC-014-3"} }
func (cfg03) Fix() string {
	return "在 harness.json 的 languages 中声明该语言，或删除 tools 中对应的项；tools 只能覆盖 format 与 lint"
}

func (cfg03) Check(s *workspace.Snapshot) []engine.Violation {
	var names []string
	for name := range s.Config.Tools {
		names = append(names, name)
	}
	sort.Strings(names)
	var out []engine.Violation
	for _, name := range names {
		if !contains(s.Config.Languages, name) {
			out = append(out, engine.Violation{File: workspace.ConfigFile, Line: 1, Msg: "tools 覆盖了未在 languages 中声明的语言 " + name})
		}
		var keys []string
		for k := range s.Config.Tools[name] {
			if k != "format" && k != "lint" {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		for _, k := range keys {
			out = append(out, engine.Violation{File: workspace.ConfigFile, Line: 1, Msg: "tools." + name + " 中的 " + k + " 不是可覆盖的命令（只能是 format、lint）"})
		}
	}
	return out
}
