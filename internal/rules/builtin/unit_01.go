package builtin

import (
	"regexp"
	"sort"
	"strings"

	"github.com/donkeydluffy/harness/internal/lang"
	"github.com/donkeydluffy/harness/internal/model"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/testreport"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// unit01：每个接口至少有一个测试名含其编号（IF-002 写作 IF_002）且通过的单元测试；匹配的测试有失败时也违规（AC-021-1）。
// 只检查路径内已有源文件的模块的接口：模块还没有实现时不要求测试。读取测试报告，与语言无关。
type unit01 struct{}

func (unit01) ID() string          { return "unit.01" }
func (unit01) Covers() []string    { return []string{"AC-021-1"} }
func (unit01) Needs() engine.Input { return engine.InputSources | engine.InputTests }
func (unit01) Fix() string {
	return "为接口编写测试名含其编号的单元测试（如 TestSave_IF_002），修复失败的测试后重新生成测试报告"
}

func (unit01) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, e := range archEntities(s, "IF") {
		mod := s.Project.ByID[e.Fields["Module"]]
		if mod == nil || !effective(mod) || !implemented(s, mod) {
			continue
		}
		re := ifTestRe(e.ID)
		var passed, failed []string
		for _, tc := range s.Tests {
			if !re.MatchString(tc.Name) || len(caseRefs(tc.Name)) > 0 {
				continue
			}
			switch tc.Status {
			case testreport.Passed:
				passed = append(passed, tc.Name)
			case testreport.Failed:
				failed = append(failed, tc.Name)
			}
		}
		switch {
		case len(failed) > 0:
			sort.Strings(failed)
			out = append(out, at(e, "%s 的单元测试在报告中为 failed：%s", e.ID, strings.Join(failed, "、")))
		case len(passed) == 0:
			out = append(out, at(e, "%s 没有测试名含 %s 且通过的单元测试", e.ID, strings.ReplaceAll(e.ID, "-", "_")))
		}
	}
	return out
}

// ifTestRe 匹配测试名中的接口编号：IF-002 匹配 TestSave_IF_002、TestSave_IF_002_Empty，不匹配 IF_0021。
func ifTestRe(id string) *regexp.Regexp {
	return regexp.MustCompile(regexp.QuoteMeta(strings.ReplaceAll(id, "-", "_")) + `([^0-9]|$)`)
}

// implemented 判断模块路径内是否有非测试源文件。
func implemented(s *workspace.Snapshot, mod *model.Entity) bool {
	for _, f := range s.Sources {
		if f.Kind == lang.NotTest && workspace.MatchAny(mod.List("Paths"), f.Path) {
			return true
		}
	}
	return false
}

// coverable 判断项目是否有可采集覆盖率的非测试源文件（Release 2 只有 Go，AC-021-2）。
func coverable(s *workspace.Snapshot) bool {
	for _, f := range s.Sources {
		if _, ok := lang.AnalyzerFor(f.Path); ok && f.Kind == lang.NotTest {
			return true
		}
	}
	return false
}

// unitReportProblem 返回 unit 报告缺失、无法读取或过期时的违规；报告可用时返回 nil。
func unitReportProblem(s *workspace.Snapshot) *engine.Violation {
	at := func(msg string) *engine.Violation {
		return &engine.Violation{File: testreport.UnitFile, Line: 1, Msg: msg}
	}
	switch {
	case s.UnitErr != nil:
		return at("unit 报告无法读取：" + s.UnitErr.Error())
	case s.Unit == nil:
		return at("没有 unit 报告；运行 harness check unit 生成")
	case s.Unit.SourceHash != s.SourceHash():
		return at("unit 报告已过期：生成报告后源码有改动；运行 harness check unit 重新生成")
	}
	return nil
}
