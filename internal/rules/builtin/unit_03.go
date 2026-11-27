package builtin

import (
	"strconv"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// unit03：Core 为 true 的模块必须设置 Coverage；模块行覆盖率不低于 Coverage（AC-021-3）。
// 项目没有可采集覆盖率的源文件时跳过（AC-021-2/B1）；报告缺失或过期由 unit.02 报告。没有语句行的模块不在报告中，不检查。
type unit03 struct{}

func (unit03) ID() string          { return "unit.03" }
func (unit03) Covers() []string    { return []string{"AC-021-3"} }
func (unit03) Needs() engine.Input { return engine.InputSources | engine.InputTests }
func (unit03) Fix() string {
	return "为 Core 模块设置 Coverage；为覆盖率不足的模块补单元测试后运行 harness check unit"
}

func (unit03) Check(s *workspace.Snapshot) []engine.Violation {
	if !coverable(s) {
		return nil
	}
	report := unitReportProblem(s) == nil
	var out []engine.Violation
	for _, m := range archEntities(s, "MOD") {
		if !effective(m) {
			continue
		}
		thr, set := m.Field("Coverage")
		if !set {
			if c, _ := m.Field("Core"); c == "true" {
				out = append(out, at(m, "%s 的 Core 为 true，但没有设置 Coverage", m.ID))
			}
			continue
		}
		want, err := strconv.ParseFloat(thr, 64)
		got, ok := 0.0, false
		if report {
			got, ok = s.Unit.Modules[m.ID]
		}
		if err == nil && ok && got < want {
			out = append(out, at(m, "%s 的行覆盖率为 %s%%，低于阈值 %s%%", m.ID, strconv.FormatFloat(got, 'f', -1, 64), thr))
		}
	}
	return out
}
