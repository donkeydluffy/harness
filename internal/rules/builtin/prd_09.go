package builtin

import (
	"fmt"
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// prd09：需求描述与验收标准中不得出现模糊词；代码格式中的文字不检查。
type prd09 struct{}

func (prd09) ID() string       { return "prd.09" }
func (prd09) Covers() []string { return []string{"AC-040-1"} }
func (prd09) Fix() string {
	return "把模糊词改为可判定的表述（具体数值、条件或枚举）；词表在 harness.json 的 vagueWords 中配置"
}

func (prd09) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, e := range requirements(s) {
		lines, nums := bodyLines(e)
		for i, l := range lines {
			for _, w := range s.Config.VagueWords {
				if w != "" && strings.Contains(l, w) {
					out = append(out, engine.Violation{File: e.File, Line: nums[i],
						Msg: fmt.Sprintf("%s 含模糊词 %q", e.ID, w)})
				}
			}
		}
	}
	return out
}
