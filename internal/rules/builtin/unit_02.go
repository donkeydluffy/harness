package builtin

import (
	"fmt"
	"sort"
	"strings"

	"github.com/donkeydluffy/harness/internal/lang"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// unit02：unit 报告存在且未过期；接口的单元测试执行了锚点所在函数的入口行（AC-021-2）。
// 类型声明没有入口行，不检查。声明的语言有源文件却不支持覆盖率采集时给出告警，说明跳过了本条与 AC-021-3。
// 只读报告，不运行外部工具（ADR-011）。
type unit02 struct{}

func (unit02) ID() string          { return "unit.02" }
func (unit02) Covers() []string    { return []string{"AC-021-2"} }
func (unit02) Needs() engine.Input { return engine.InputSources | engine.InputTests }
func (unit02) Fix() string {
	return "让接口的单元测试真正调用锚点所在的函数，然后运行 harness check unit 重新生成报告"
}

func (unit02) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, name := range uncoverable(s) {
		out = append(out, engine.Violation{File: workspace.ConfigFile, Line: 1, Warn: true,
			Msg: name + " 不支持覆盖率采集，跳过 AC-021-2 与 AC-021-3"})
	}
	if !coverable(s) {
		return out
	}
	if v := unitReportProblem(s); v != nil {
		return append(out, *v)
	}
	for _, d := range s.Anchors() {
		cov, ok := s.Unit.Interfaces[d.ID]
		if !ok || d.EntryLine == 0 || len(cov.Tests) == 0 || cov.EntryHit {
			continue
		}
		out = append(out, engine.Violation{File: d.File, Line: d.Line,
			Msg: fmt.Sprintf("%s 的单元测试 %s 没有调用 %s：入口行 %s:%d 未执行", d.ID, strings.Join(cov.Tests, "、"), d.ID, d.File, d.EntryLine)})
	}
	return out
}

// uncoverable 返回声明的、有源文件却没有 Analyzer 的语言（升序）。
func uncoverable(s *workspace.Snapshot) []string {
	has, ok := map[string]bool{}, map[string]bool{}
	for _, f := range s.Sources {
		if f.Lang == nil {
			continue
		}
		has[f.Lang.Name()] = true
		if _, a := lang.AnalyzerFor(f.Path); a {
			ok[f.Lang.Name()] = true
		}
	}
	var out []string
	for _, name := range s.Config.Languages {
		if has[name] && !ok[name] {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}
