package builtin

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/donkeydluffy/harness/internal/model"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/testreport"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// arch05：每个扩展点至少有一个测试名含其编号、状态为 passed 的示例扩展测试；
// 关联的扩展场景都不在当前发布范围时不检查。
type arch05 struct{}

func (arch05) Needs() engine.Input { return engine.InputTests }
func (arch05) ID() string          { return "arch.05" }
func (arch05) Covers() []string    { return []string{"AC-033-4"} }
func (arch05) Fix() string {
	return "为扩展点编写示例扩展测试，测试名含编号（如 TestExample_EXT_01），运行后把报告写到 harness.json 的 testReports 路径"
}

var extTokenRe = regexp.MustCompile(`EXT[-_](\d+)`)

func (arch05) Check(s *workspace.Snapshot) []engine.Violation {
	var exts []*model.Entity
	for _, e := range s.Project.ByKind["EXT"] {
		if !workspace.Removed(e) && extInScope(s, e) {
			exts = append(exts, e)
		}
	}
	if len(exts) == 0 {
		return nil
	}
	if len(s.Reports) == 0 {
		return []engine.Violation{{File: ".", Line: 1, Msg: "输入缺失：没有测试报告（harness.json 的 testReports：" + strings.Join(s.Config.TestReports, ", ") + "）"}}
	}
	passed := map[int]bool{}
	for _, tc := range s.Tests {
		if tc.Status != testreport.Passed {
			continue
		}
		for _, m := range extTokenRe.FindAllStringSubmatch(tc.Name, -1) {
			n, _ := strconv.Atoi(m[1])
			passed[n] = true
		}
	}
	var out []engine.Violation
	for _, e := range exts {
		n, _ := strconv.Atoi(strings.TrimPrefix(e.ID, "EXT-"))
		if !passed[n] {
			out = append(out, at(e, "扩展点 %s 没有通过的示例扩展测试（测试名需含 %s）", e.ID, strings.ReplaceAll(e.ID, "-", "_")))
		}
	}
	return out
}

// extInScope：关联的需求中至少一个存在且在当前发布范围内。没有关联或关联不存在时
// 由 arch.04、trace.03 报告，这里不重复。
func extInScope(s *workspace.Snapshot, e *model.Entity) bool {
	for _, r := range e.List("Relates") {
		if t := s.Project.ByID[r]; t != nil && inScope(s, t) {
			return true
		}
	}
	return false
}
